package main

import (
	"os"
	"os/signal"
	"syscall"

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
// @BasePath /api
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

	// Initialize gRPC server (optional)
	grpcSrv := initializer.InitGRPCServer(cfg, services, logger)
	if grpcSrv != nil {
		defer grpcSrv.Stop()
	}

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info(logging.General, logging.Startup, "Shutting down gracefully...", nil)

		// Cleanup will happen via defer statements
		os.Exit(0)
	}()

	// Create app context
	appCtx := &api.AppContext{
		DB:                  db,
		Logger:              logger,
		Config:              cfg,
		NotificationService: services.Notification,
		WebSocketHub:        services.WebSocketHub,
		AuthClient:          services.AuthClient,
	}

	// Start HTTP server (blocks until shutdown)
	api.InitServer(appCtx)
}
