package api

import (
	"fmt"
	"time"

	"github.com/bytedance/sonic"
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
	WebSocketHub        *wsHub.Hub
	AuthClient          *auth.Client // Auth client for service token validation
}

func InitServer(ctx *AppContext) {
	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:     ctx.Config.Server.Name,
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	// Middleware
	app.Use(middleware.DefaultStructuredLogger(&ctx.Config.Logger))

	// Security middleware
	app.Use(middleware.SecurityHeaders(middleware.DefaultSecurityHeadersConfig()))
	app.Use(middleware.RequestValidation(middleware.DefaultRequestValidationConfig()))

	app.Use(middleware.Prometheus())
	app.Use(middleware.Tracing(middleware.TracingConfig{
		ServiceName: "notifier-service",
	}))
	app.Use(middleware.Cors(ctx.Config.Cors.AllowOrigins))
	app.Use(recover.New())

	// Tenant middleware - extract and validate tenant context
	app.Use(middleware.TenantMiddleware(middleware.TenantConfig{
		Enabled:            true,
		HeaderName:         "X-Tenant-ID",
		AllowMissingTenant: true, // Allow missing for public routes
		ContextKey:         "tenantId",
		SkipPaths:          []string{"/health", "/ready", "/swagger", "/metrics", "/ws"},
		TenantValidator: func(tenantID string) bool {
			// Validate tenant exists and is active
			if ctx.DB == nil {
				return true // Skip validation if DB not available
			}
			var count int64
			ctx.DB.Table("tenants").Where("id = ? AND is_active = ?", tenantID, true).Count(&count)
			return count > 0
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

	// Metrics endpoint
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Start the server
	ctx.Logger.Info(logging.General, logging.Startup, "Server started", map[logging.ExtraKey]interface{}{
		"port": ctx.Config.Server.InternalPort,
	})

	err := app.Listen(fmt.Sprintf(":%s", ctx.Config.Server.InternalPort))
	if err != nil {
		ctx.Logger.Fatal(logging.General, logging.Startup, err.Error(), nil)
	}
}

func RegisterRoutes(app *fiber.App, ctx *AppContext) {
	// Create repositories
	templateRepo := repository.NewNotificationTemplateRepository(ctx.DB, ctx.Logger)
	prefRepo := repository.NewNotificationPreferenceRepository(ctx.DB, ctx.Logger)

	// Create handlers
	notificationHandler := handlers.NewNotificationHandler(ctx.NotificationService)
	preferenceHandler := handlers.NewPreferenceHandler(prefRepo)
	templateHandler := handlers.NewTemplateHandler(templateRepo)

	// Create an API group
	api := app.Group("/api")

	// Create a v1 group
	v1 := api.Group("/v1")
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

			// User JWT auth middleware for direct API access
			authConfig := middleware.AuthConfig{
				Enabled:   true,
				Secret:    ctx.Config.Auth.JWTSecret,
				SkipPaths: []string{"/api/v1/health", "/api/v1/sms"},
			}
			jwtMiddleware := middleware.AuthMiddleware(authConfig)

			// Service token auth middleware for service-to-service communication
			var serviceMiddlewareHandler fiber.Handler
			if ctx.AuthClient != nil {
				serviceAuthConfig := serviceMiddleware.ServiceAuthConfig{
					AuthClient: ctx.AuthClient,
					Logger:     ctx.Logger,
					CacheTTL:   5 * time.Minute,
					SkipPaths:  []string{"/api/v1/health"},
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

			// Protected template routes (require user JWT + admin role or templates:manage permission)
			templates := v1.Group("/templates", jwtMiddleware, middleware.RequirePermissions("templates:read"))
			routers.Templates(templates, templateHandler)

		} else {
			ctx.Logger.Warn(logging.General, logging.Startup, "Authentication is DISABLED - all routes are public", nil)

			// When auth is disabled, all routes are public (NOT recommended for production)
			notifications := v1.Group("/notifications")
			routers.Notifications(notifications, notificationHandler)

			preferences := v1.Group("/preferences")
			routers.Preferences(preferences, preferenceHandler)

			templates := v1.Group("/templates")
			routers.Templates(templates, templateHandler)
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
