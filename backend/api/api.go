package api

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/gofiber/websocket/v2"
	"github.com/minisource/go-common/http/middleware"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-common/response"
	"github.com/minisource/go-sdk/auth"
	serviceMiddleware "github.com/minisource/notifier/api/middleware"
	"github.com/minisource/notifier/api/v1/handlers"
	routers "github.com/minisource/notifier/api/v1/routes"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/retention"
	"github.com/minisource/notifier/internal/service"
	wsHub "github.com/minisource/notifier/internal/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

type AppContext struct {
	DB                  *gorm.DB
	Logger              logging.Logger
	Config              *config.Config
	NotificationService *service.NotificationService
	TemplateService     *service.TemplateService
	PreferenceService   *service.PreferenceService
	ReminderService     *service.ReminderService
	ProviderRepo        repository.ProviderRepository
	SettingRepo         repository.SettingRepository
	TenantRepo          repository.TenantRepository
	WebSocketHub        *wsHub.Hub
	AuthClient          *auth.Client // Auth client for service token validation
	BalanceService      *service.BalanceService
	BalanceRepo         repository.ProviderBalanceRepository
	DeliveryControl     *service.DeliveryControlService
	RetentionScheduler  *retention.Scheduler
}

func InitServer(ctx *AppContext) *fiber.App {
	// Validate config at startup
	if issues := ctx.Config.Validate(); len(issues) > 0 {
		for _, issue := range issues {
			log.Printf("[CONFIG WARNING] %s", issue)
		}
		// Fail in production if critical config is missing
		runMode := ctx.Config.Server.RunMode
		if runMode == "production" || runMode == "staging" {
			for _, issue := range issues {
				log.Fatalf("[CONFIG ERROR] %s", issue)
			}
		}
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:     ctx.Config.Server.Name,
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Middleware — Request ID FIRST (before any other middleware)
	app.Use(middleware.RequestID())

	// Security middleware
	app.Use(middleware.SecurityHeaders(middleware.DefaultSecurityHeadersConfig()))
	app.Use(middleware.RequestValidation(middleware.DefaultRequestValidationConfig()))

	// Metrics and tracing
	app.Use(middleware.Prometheus())
	app.Use(middleware.Tracing(middleware.TracingConfig{
		ServiceName: "notifier-service",
	}))

	// Structured access logging — canonical http_request_completed event.
	// Runs AFTER tracing so the span exists when the event is emitted.
	// Safe metadata only; body capture is opt-in via OBSERVABILITY_* env vars.
	app.Use(middleware.AccessLog(middleware.LoadAccessLogConfigFromEnv("notifier-service", ctx.Logger)))

	app.Use(middleware.Cors(ctx.Config.Cors.AllowOrigins))

	// OPTIONS fast path — after CORS (headers are set), before expensive middleware
	// This ensures preflight requests skip auth/DB/rate-limit/tenant for maximum speed.
	app.Use(func(c *fiber.Ctx) error {
		if c.Method() == "OPTIONS" {
			c.Set("Access-Control-Max-Age", "86400")
			return c.SendStatus(204)
		}
		return c.Next()
	})

	// Rate limiting — controlled by RATE_LIMIT_ENABLED env var (default: true).
	// A shared limiter instance is used so route-specific (stricter) limits can
	// be registered BEFORE app.Use: high-risk delivery-control mutations get a
	// low-frequency limit, while status/history/held polling gets a bounded
	// limit (prevents abuse/DoS of the control plane).
	rateLimiter := serviceMiddleware.NewRateLimiter(serviceMiddleware.RateLimitConfig{
		Enabled:  ctx.Config.RateLimit.Enabled,
		Requests: ctx.Config.RateLimit.Requests,
		Window:   time.Duration(ctx.Config.RateLimit.WindowSeconds) * time.Second,
	})
	if ctx.Config.RateLimit.Enabled {
		if m := ctx.Config.RateLimit.ControlMutationRequests; m > 0 {
			rateLimiter.SetRouteLimit("/v1/admin/delivery-control/pause", m, time.Minute)
			rateLimiter.SetRouteLimit("/v1/admin/delivery-control/resume", m, time.Minute)
		}
		if r := ctx.Config.RateLimit.ControlReadRequests; r > 0 {
			rateLimiter.SetRouteLimit("/v1/admin/delivery-control/status", r, time.Minute)
			rateLimiter.SetRouteLimit("/v1/admin/delivery-control/history", r, time.Minute)
			rateLimiter.SetRouteLimit("/v1/admin/delivery-control/held", r, time.Minute)
		}
	}
	app.Use(rateLimiter.Handler())

	// Request timing middleware — measures request duration, logs SLOW_REQUEST for requests >1s
	app.Use(serviceMiddleware.RequestTimingMiddleware(serviceMiddleware.RequestTimingConfig{
		SlowRequestThreshold: 1 * time.Second,
		Logger:               ctx.Logger,
	}))

	// Panic recovery (last safety net)
	app.Use(recover.New())

	// Tenant middleware - extract and validate tenant context
	app.Use(middleware.TenantMiddleware(middleware.TenantConfig{
		Enabled:            true,
		HeaderName:         "X-Tenant-ID",
		AllowMissingTenant: true, // Allow missing for public routes
		ContextKey:         "tenantId",
		SkipPaths:          []string{"/health", "/ready", "/swagger", "/metrics", "/ws"},
		TenantValidator: func(tenantID string) bool {
			// Tenants are owned by the Auth service. Membership + active status
			// are enforced there (via the user's JWT and the /users/me/tenants
			// list the frontend selects from). Notifier no longer owns a
			// tenants table, so any tenant id supplied by an authenticated
			// caller is accepted and used only as a scoping key.
			return true
		},
	}))

	// WebSocket upgrade middleware for /ws routes
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Register routes
	RegisterRoutes(app, ctx)

	// Route dump for debugging in development mode
	if ctx.Config.Server.RunMode == "development" {
		log.Println("[ROUTE DUMP] Registered routes:")
		for _, route := range app.GetRoutes() {
			log.Printf("[ROUTE] %s %s", route.Method, route.Path)
		}
	}

	// Health and readiness endpoints
	app.Get("/health", func(c *fiber.Ctx) error {
		return response.New().Data(fiber.Map{"status": "healthy", "service": "notifier"}).Send(c)
	})
	app.Get("/ready", func(c *fiber.Ctx) error {
		return response.New().Data(fiber.Map{"status": "ready"}).Send(c)
	})

	// Metrics endpoint
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Start the server
	ctx.Logger.Info(logging.General, logging.Startup, "Server starting", map[logging.ExtraKey]interface{}{
		"port": ctx.Config.Server.InternalPort,
	})

	return app
}

func RegisterRoutes(app *fiber.App, ctx *AppContext) {
	// Create handlers
	notificationHandler := handlers.NewNotificationHandler(ctx.NotificationService)
	preferenceHandler := handlers.NewPreferenceHandler(ctx.PreferenceService)
	templateHandler := handlers.NewTemplateHandler(ctx.TemplateService)
	inAppHandler := handlers.NewInAppHandler(ctx.NotificationService.GetRepository())
	reminderHandler := handlers.NewReminderHandler(ctx.ReminderService)
	providerHandler := handlers.NewProviderHandler(ctx.ProviderRepo, ctx.NotificationService.GetRepository(), ctx.Logger, ctx.DeliveryControl)
	dashboardHandler := handlers.NewDashboardHandler(ctx.NotificationService)
	observabilityHandler := handlers.NewObservabilityHandler(ctx.NotificationService)
	deliveryHandler := handlers.NewDeliveryHandler(ctx.NotificationService)
	providerAttemptHandler := handlers.NewProviderAttemptHandler(ctx.NotificationService)
	settingsHandler := handlers.NewSettingsHandler(ctx.SettingRepo)
	retentionHandler := handlers.NewAdminRetentionHandler(
		retention.NewPolicyRepository(ctx.DB, ctx.Logger),
		retention.NewRunRepository(ctx.DB, ctx.Logger),
		ctx.RetentionScheduler,
		ctx.Logger,
	)
	balanceHandler := handlers.NewBalanceHandler(
		ctx.BalanceService,
		ctx.ProviderRepo,
		ctx.BalanceRepo,
		ctx.Logger,
	)
	deliveryControlHandler := handlers.NewDeliveryControlHandler(
		ctx.DeliveryControl,
		ctx.NotificationService.GetRepository(),
	)

	// Create access-aware handlers
	meHandler := handlers.NewMeHandler(
		ctx.NotificationService,
		ctx.PreferenceService,
		ctx.ReminderService,
	)
	adminHandler := handlers.NewAdminHandler(
		ctx.NotificationService,
		ctx.TemplateService,
		ctx.PreferenceService,
		ctx.ReminderService,
	)

	// Add legacy /api/v1 → /v1 rewrite middleware (for backward compatibility)
	// Only rewrites /api/v1/* paths, not bare /api/*
	app.Use("/api", func(c *fiber.Ctx) error {
		path := c.Path()
		if len(path) >= 8 && path[:8] == "/api/v1/" {
			c.Path("/v1/" + path[8:])
		} else if path == "/api/v1" {
			c.Path("/v1")
		}
		return c.Next()
	})

	// Create the canonical v1 group (without /api prefix)
	v1 := app.Group("/v1")
	{
		// Health routes (no auth required - public)
		health := v1.Group("/health")
		routers.Health(health)

		// SMS routes (legacy - public for backward compatibility)
		sms := v1.Group("/sms")
		routers.SMS(sms, ctx.Config)

		// Check if auth is enabled
		if ctx.Config.Auth.Enabled {
			ctx.Logger.Info(logging.General, logging.Startup, "Authentication is ENABLED", nil)

			// Build real auth middleware (supports JWKS / introspection / HS256)
			jwtMiddleware := serviceMiddleware.RealAuthMiddleware(serviceMiddleware.RealAuthConfig{
				ValidationMode:   ctx.Config.Auth.ValidationMode,
				HS256Secret:      ctx.Config.Auth.JWTSecret,
				JWKSURL:          ctx.Config.Auth.JWKSURL,
				JWKSCacheTTL:     time.Duration(ctx.Config.Auth.JWKSCacheTTL) * time.Second,
				JWKSRefreshOnKidMiss: true,
				IntrospectionURL: ctx.Config.Auth.IntrospectionURL,
				IntrospectionTimeout: 3 * time.Second,
				Issuer:           ctx.Config.Auth.Issuer,
				Audience:         ctx.Config.Auth.Audience,
				SkipPaths:        []string{"/v1/health", "/v1/sms"},
				RequireAuth:      true,
				AppEnv:           ctx.Config.Server.RunMode,
			})

			// Service token auth middleware for service-to-service communication
			var serviceMiddlewareHandler fiber.Handler
			if ctx.AuthClient != nil {
				serviceAuthConfig := serviceMiddleware.ServiceAuthConfig{
					AuthClient: ctx.AuthClient,
					Logger:     ctx.Logger,
					CacheTTL:   5 * time.Minute,
					SkipPaths:  []string{"/v1/health"},
				}
				serviceMiddlewareHandler = serviceMiddleware.ServiceAuthMiddleware(serviceAuthConfig)

				// Service-only routes (require service token with specific scopes)
				serviceGroup := v1.Group("/service", serviceMiddlewareHandler)
				{
					// Service notification routes (require notifications:send scope)
					serviceNotifications := serviceGroup.Group("/notifications", serviceMiddleware.RequireScope("notifications:send"))
					routers.Notifications(serviceNotifications, notificationHandler)
				}
			}

			// Protected notification routes (require user JWT)
			notifications := v1.Group("/notifications", jwtMiddleware)
			routers.Notifications(notifications, notificationHandler)

			// Protected preference routes (require user JWT)
			preferences := v1.Group("/preferences", jwtMiddleware)
			routers.Preferences(preferences, preferenceHandler)

			// Protected template routes (admin or notifier:templates:manage permission)
			templates := v1.Group("/templates", jwtMiddleware, serviceMiddleware.RequireAdminOrPermission("notifier:templates:manage"))
			routers.Templates(templates, templateHandler)

			// Protected in-app notification routes (require user JWT)
			inapp := v1.Group("/inapp", jwtMiddleware)
			routers.InApp(inapp, inAppHandler)

			// Protected reminder routes (require user JWT)
			reminders := v1.Group("/reminders", jwtMiddleware)
			routers.Reminders(reminders, reminderHandler)

			// Protected provider routes (require user JWT)
			providers := v1.Group("/providers", jwtMiddleware)
			routers.Providers(providers, providerHandler)

			// Protected dashboard routes (require user JWT)
			dashboard := v1.Group("/dashboard", jwtMiddleware)
			routers.Dashboard(dashboard, dashboardHandler)

			// Protected observability routes (require user JWT)
			observability := v1.Group("/observability", jwtMiddleware)
			routers.Observability(observability, observabilityHandler)

			// Protected delivery routes (require user JWT)
			deliveries := v1.Group("/deliveries", jwtMiddleware)
			routers.Deliveries(deliveries, deliveryHandler)

			// User API — /me routes (userId from JWT, never from path/body)
			me := v1.Group("/me", jwtMiddleware)
			routers.Me(me, meHandler)

			// Admin API — /admin routes (require admin or notifier:admin permission)
			admin := v1.Group("/admin", jwtMiddleware, serviceMiddleware.RequireAdmin())

			// NOTE: static admin sub-routes MUST be registered before routers.Admin() —
			// Fiber's router gives priority to the first-registered route when a static
			// segment (e.g. /tenants) collides with a param route (e.g. /notifications/:notificationId).

			// Admin dashboard overview (uses DashboardHandler — needs admin role)
			adminDashboard := admin.Group("/notifications/dashboard")
			adminDashboard.Get("/overview", dashboardHandler.GetDashboardOverview)

			// Admin settings routes (namespaced to avoid collision with auth /v1/admin/settings)
			adminSettings := admin.Group("/notifications/settings")
			routers.Settings(adminSettings, settingsHandler)

			// Admin log retention
			adminRetention := admin.Group("/log-retention")
			{
				adminRetention.Get("/categories", retentionHandler.ListCategories)
				adminRetention.Get("/policies", retentionHandler.ListPolicies)
				adminRetention.Get("/policies/:id", retentionHandler.GetPolicy)
				adminRetention.Put("/policies/:id", retentionHandler.UpsertPolicy)
				adminRetention.Post("/policies/:id/enable", retentionHandler.EnablePolicy)
				adminRetention.Post("/policies/:id/disable", retentionHandler.DisablePolicy)
				adminRetention.Post("/policies/:id/preview", retentionHandler.PreviewCleanup)
				adminRetention.Post("/policies/:id/run", retentionHandler.RunCleanup)
				adminRetention.Get("/runs", retentionHandler.ListRuns)
				adminRetention.Get("/runs/:id", retentionHandler.GetRun)
			}

			routers.Admin(admin, adminHandler)

			// Admin global outbound delivery pause / emergency freeze routes
			adminDeliveryControl := admin.Group("/delivery-control")
			routers.DeliveryControl(adminDeliveryControl, deliveryControlHandler)

			// Admin provider balance / quota monitoring routes — static /balance
			// group MUST be registered before adminProviders (whose :providerId
			// param route would otherwise swallow GET /providers/balance).
			adminBalance := admin.Group("/providers/balance")
			routers.ProviderBalance(adminBalance, balanceHandler)
			admin.Get("/providers/:providerId/balance", balanceHandler.GetHealth)
			admin.Post("/providers/:providerId/balance/refresh", balanceHandler.Refresh)
			admin.Put("/providers/:providerId/balance/settings", balanceHandler.UpdateSettings)
			admin.Get("/providers/:providerId/balance/alerts", balanceHandler.ListAlerts)

			// Admin provider routes — use ProviderHandler (separate from AdminHandler)
			adminProviders := admin.Group("/providers")
			routers.Providers(adminProviders, providerHandler)

			// Admin delivery routes
			adminDeliveries := admin.Group("/deliveries")
			routers.Deliveries(adminDeliveries, deliveryHandler)

			// Admin provider attempt (provider request lifecycle logging) routes
			adminAttempts := admin.Group("/attempts")
			routers.ProviderAttempts(adminAttempts, providerAttemptHandler)
			admin.Get("/notifications/:notificationId/attempts", providerAttemptHandler.ListNotificationAttempts)

			// Admin observability routes
			adminObservability := admin.Group("/observability")
			routers.Observability(adminObservability, observabilityHandler)
		} else {
			ctx.Logger.Warn(logging.General, logging.Startup, "Authentication is DISABLED - all routes are public", nil)

			// When auth is disabled, all routes are public (NOT recommended for production)
			notifications := v1.Group("/notifications")
			routers.Notifications(notifications, notificationHandler)

			preferences := v1.Group("/preferences")
			routers.Preferences(preferences, preferenceHandler)

			templates := v1.Group("/templates")
			routers.Templates(templates, templateHandler)

			inapp := v1.Group("/inapp")
			routers.InApp(inapp, inAppHandler)

			reminders := v1.Group("/reminders")
			routers.Reminders(reminders, reminderHandler)

			providers := v1.Group("/providers")
			routers.Providers(providers, providerHandler)

			dashboard := v1.Group("/dashboard")
			routers.Dashboard(dashboard, dashboardHandler)

			observability := v1.Group("/observability")
			routers.Observability(observability, observabilityHandler)

			deliveries := v1.Group("/deliveries")
			routers.Deliveries(deliveries, deliveryHandler)

			// User API — /me routes (userId from JWT, no auth check)
			me := v1.Group("/me")
			routers.Me(me, meHandler)

			// Admin API — /admin routes (no auth check in dev mode)
			admin := v1.Group("/admin")

			// NOTE: static admin sub-routes MUST be registered before routers.Admin() —
			// Fiber's router gives priority to the first-registered route when a static
			// segment (e.g. /tenants) collides with a param route (e.g. /notifications/:notificationId).

			// Admin dashboard overview
			adminDashboard := admin.Group("/notifications/dashboard")
			adminDashboard.Get("/overview", dashboardHandler.GetDashboardOverview)

			// Admin settings routes (namespaced to avoid collision with auth /v1/admin/settings)
			adminSettings := admin.Group("/notifications/settings")
			routers.Settings(adminSettings, settingsHandler)

			// Admin log retention
			adminRetention := admin.Group("/log-retention")
			{
				adminRetention.Get("/categories", retentionHandler.ListCategories)
				adminRetention.Get("/policies", retentionHandler.ListPolicies)
				adminRetention.Get("/policies/:id", retentionHandler.GetPolicy)
				adminRetention.Put("/policies/:id", retentionHandler.UpsertPolicy)
				adminRetention.Post("/policies/:id/enable", retentionHandler.EnablePolicy)
				adminRetention.Post("/policies/:id/disable", retentionHandler.DisablePolicy)
				adminRetention.Post("/policies/:id/preview", retentionHandler.PreviewCleanup)
				adminRetention.Post("/policies/:id/run", retentionHandler.RunCleanup)
				adminRetention.Get("/runs", retentionHandler.ListRuns)
				adminRetention.Get("/runs/:id", retentionHandler.GetRun)
			}

			routers.Admin(admin, adminHandler)

			// Admin global outbound delivery pause / emergency freeze routes
			adminDeliveryControl := admin.Group("/delivery-control")
			routers.DeliveryControl(adminDeliveryControl, deliveryControlHandler)

			// Admin provider balance / quota monitoring routes — static /balance
			// group MUST be registered before adminProviders (whose :providerId
			// param route would otherwise swallow GET /providers/balance).
			adminBalance := admin.Group("/providers/balance")
			routers.ProviderBalance(adminBalance, balanceHandler)
			admin.Get("/providers/:providerId/balance", balanceHandler.GetHealth)
			admin.Post("/providers/:providerId/balance/refresh", balanceHandler.Refresh)
			admin.Put("/providers/:providerId/balance/settings", balanceHandler.UpdateSettings)
			admin.Get("/providers/:providerId/balance/alerts", balanceHandler.ListAlerts)

			// Admin provider routes
			adminProviders := admin.Group("/providers")
			routers.Providers(adminProviders, providerHandler)

			// Admin delivery routes
			adminDeliveries := admin.Group("/deliveries")
			routers.Deliveries(adminDeliveries, deliveryHandler)

			// Admin provider attempt (provider request lifecycle logging) routes
			adminAttempts := admin.Group("/attempts")
			routers.ProviderAttempts(adminAttempts, providerAttemptHandler)
			admin.Get("/notifications/:notificationId/attempts", providerAttemptHandler.ListNotificationAttempts)

			// Admin observability routes
			adminObservability := admin.Group("/observability")
			routers.Observability(adminObservability, observabilityHandler)
		}
	}

	// WebSocket endpoint with authentication
	ws := app.Group("/ws")
	{
		// Add WebSocket authentication middleware
		if ctx.Config.Auth.Enabled && ctx.AuthClient != nil {
			ws.Use(func(c *fiber.Ctx) error {
				// Get token from query string for WebSocket upgrade
				token := c.Query("token")
				if token == "" {
					// Also check Authorization header
					authHeader := c.Get("Authorization")
					if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
						token = authHeader[7:]
					}
				}

				if token == "" {
					ctx.Logger.Warn(logging.General, logging.Api, "WebSocket connection rejected: missing token", nil)
					return response.Unauthorized(c, "Missing authentication token")
				}

				// Validate service token using auth client
				validation, err := ctx.AuthClient.ValidateToken(c.Context(), token)
				if err != nil || !validation.Valid {
					ctx.Logger.Warn(logging.General, logging.Api, "WebSocket connection rejected: invalid token", map[logging.ExtraKey]interface{}{
						"error": err,
					})
					return response.Unauthorized(c, "Invalid token")
				}

				// Store service info in context
				c.Locals("clientId", validation.ClientID)
				c.Locals("serviceName", validation.ServiceName)
				c.Locals("scopes", validation.Scopes)

				ctx.Logger.Debug(logging.General, logging.Api, "WebSocket connection authorized", map[logging.ExtraKey]interface{}{
					"clientId":    validation.ClientID,
					"serviceName": validation.ServiceName,
				})

				return c.Next()
			})
		}
		routers.WebSocket(ws, ctx.WebSocketHub)
	}

	ctx.Logger.Info(logging.General, logging.Startup, "Routes registered successfully", nil)
}
