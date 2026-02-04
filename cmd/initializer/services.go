package initializer

import (
	"time"

	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-sdk/auth"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/service"
	"github.com/minisource/notifier/internal/websocket"
	"github.com/minisource/notifier/internal/worker"
)

// Services holds all service instances
type Services struct {
	Notification *service.NotificationService
	Template     *service.TemplateService
	Preference   *service.PreferenceService
	Worker       *worker.NotificationWorker
	WebSocketHub *websocket.Hub
	AuthClient   *auth.Client
}

// InitServices creates all service instances
func InitServices(cfg *config.Config, repos *Repositories, logger logging.Logger) *Services {
	// Initialize WebSocket hub
	logger.Info(logging.General, logging.Startup, "Initializing WebSocket hub", nil)
	wsHub := websocket.NewHub(logger)
	wsHub.Start()

	// Initialize notification service (without worker initially)
	logger.Info(logging.General, logging.Startup, "Initializing notification service", nil)
	notificationService := service.NewNotificationService(
		cfg,
		logger,
		repos.Notification,
		repos.Template,
		repos.Preference,
		repos.Log,
		repos.Setting,
		repos.SMSTemplate,
		nil, // Worker will be set after initialization
		wsHub,
	)

	// Initialize template service
	logger.Info(logging.General, logging.Startup, "Initializing template service", nil)
	templateService := service.NewTemplateService(repos.Template, logger)

	// Initialize preference service
	logger.Info(logging.General, logging.Startup, "Initializing preference service", nil)
	preferenceService := service.NewPreferenceService(repos.Preference, logger)

	// Create handler adapters
	smsHandler := service.NewSMSHandlerAdapter(notificationService)
	emailHandler := service.NewEmailHandlerAdapter(notificationService)
	pushHandler := service.NewPushHandlerAdapter(notificationService)

	// Initialize worker
	logger.Info(logging.General, logging.Startup, "Initializing notification worker", map[logging.ExtraKey]interface{}{
		"numWorkers": cfg.Worker.NumWorkers,
		"queueSize":  cfg.Worker.QueueSize,
	})
	notificationWorker := worker.NewNotificationWorker(
		cfg,
		logger,
		repos.Notification,
		repos.Log,
		smsHandler,
		emailHandler,
		pushHandler,
	)
	notificationWorker.Start()

	// Update service with worker reference
	notificationService = service.NewNotificationService(
		cfg,
		logger,
		repos.Notification,
		repos.Template,
		repos.Preference,
		repos.Log,
		repos.Setting,
		repos.SMSTemplate,
		notificationWorker,
		wsHub,
	)

	// Initialize auth client (optional)
	authClient := initAuthClient(cfg, logger)

	return &Services{
		Notification: notificationService,
		Template:     templateService,
		Preference:   preferenceService,
		Worker:       notificationWorker,
		WebSocketHub: wsHub,
		AuthClient:   authClient,
	}
}

// initAuthClient creates auth client if enabled
func initAuthClient(cfg *config.Config, logger logging.Logger) *auth.Client {
	if !cfg.Auth.Enabled || cfg.Auth.BaseURL == "" {
		return nil
	}

	logger.Info(logging.General, logging.Startup, "Initializing auth client", map[logging.ExtraKey]interface{}{
		"baseURL": cfg.Auth.BaseURL,
	})

	return auth.NewClient(auth.ClientConfig{
		BaseURL:      cfg.Auth.BaseURL,
		ClientID:     cfg.Auth.ClientID,
		ClientSecret: cfg.Auth.ClientSecret,
		Timeout:      10 * time.Second,
		AutoRefresh:  true,
	})
}
