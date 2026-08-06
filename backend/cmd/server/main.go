package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/api"
	"github.com/minisource/notifier/cmd/initializer"
	_ "github.com/minisource/notifier/docs" // Import swagger docs
)

// @title Notifier Service API
// @version 2.0
// @description Notification Service for Minisource - Handles Email, SMS, Push, and In-App notifications
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@minisource.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host 127.0.0.1:9002
// @BasePath /v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Initialize configuration
	cfg := initializer.InitConfig()

	// Initialize logger
	logger := initializer.InitLogger(cfg)

	// Initialize metrics
	initializer.InitMetrics()

	// Initialize tracing (optional)
	tp := initializer.InitTracing(cfg, logger)
	if tp != nil {
		defer initializer.ShutdownTracing(tp, logger)
	}

	// Initialize translator
	initializer.InitTranslator(logger)

	// Initialize database
	db := initializer.InitDatabase(cfg, logger)
	defer initializer.CloseDatabase(db, logger)

	// Initialize repositories
	repos := initializer.InitRepositories(db, logger)

	// Initialize services (includes WebSocket hub and worker)
	services := initializer.InitServices(cfg, repos, logger)
	defer services.WebSocketHub.Stop()
	defer services.Worker.Stop()

	// Start the provider balance scheduler (account balance/quota monitoring)
	services.BalanceScheduler.Start()
	defer services.BalanceScheduler.Stop()

	// Initialize gRPC server (optional)
	grpcSrv := initializer.InitGRPCServer(cfg, services, logger)
	if grpcSrv != nil {
		defer grpcSrv.Stop()
	}

	// Create app context
	appCtx := &api.AppContext{
		DB:                  db,
		Logger:              logger,
		Config:              cfg,
		NotificationService: services.Notification,
		TemplateService:     services.Template,
		PreferenceService:   services.Preference,
		ReminderService:     services.Reminder,
		ProviderRepo:        repos.Provider,
		SettingRepo:         repos.Setting,
		TenantRepo:          repos.Tenant,
		WebSocketHub:        services.WebSocketHub,
		AuthClient:          services.AuthClient,
		BalanceService:      services.Balance,
		BalanceRepo:         repos.ProviderBalance,
		DeliveryControl:     services.DeliveryControl,
		RetentionScheduler:  services.RetentionScheduler,
	}

	// Initialize HTTP server (routes + middleware, no listen yet)
	app := api.InitServer(appCtx)

	// Start HTTP server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.InternalPort)
		logger.Info(logging.General, logging.Startup, "Server listening", map[logging.ExtraKey]interface{}{
			"address": addr,
		})
		if err := app.Listen(addr); err != nil && err != http.ErrServerClosed {
			logger.Fatal(logging.General, logging.Startup, "Failed to start server", map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(logging.General, logging.Startup, "Shutting down server...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error(logging.General, logging.Startup, "Server forced to shutdown", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info(logging.General, logging.Startup, "Server exited properly", nil)
}
