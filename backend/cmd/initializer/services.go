package initializer

import (
	"time"

	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-sdk/auth"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/retention"
	"github.com/minisource/notifier/internal/service"
	"gorm.io/gorm"
	"github.com/minisource/notifier/internal/websocket"
	"github.com/minisource/notifier/internal/worker"
)// Services holds all service instances
type Services struct {
	Notification     *service.NotificationService
	Template         *service.TemplateService
	Preference       *service.PreferenceService
	Reminder         *service.ReminderService
	Worker           *worker.NotificationWorker
	WebSocketHub     *websocket.Hub
	AuthClient       *auth.Client
	Balance          *service.BalanceService
	BalanceScheduler *service.BalanceScheduler
	DeliveryControl  *service.DeliveryControlService
	RetentionScheduler *retention.Scheduler
}

// InitServices creates all service instances
func InitServices(cfg *config.Config, repos *Repositories, logger logging.Logger) *Services {
	// Initialize auth client (optional)
	authClient := initAuthClient(cfg, logger)

	// Initialize WebSocket hub
	logger.Info(logging.General, logging.Startup, "Initializing WebSocket hub", nil)
	wsHub := websocket.NewHub(logger)
	wsHub.Start()

	// Initialize the global outbound delivery pause service FIRST so the
	// handler adapters (final provider-boundary gate) get a service instance
	// that actually has delivery control wired in. Creating the adapters from
	// a deliveryControl==nil service would silently disable the pause gate
	// for digest sends and the sync path.
	logger.Info(logging.General, logging.Startup, "Initializing delivery control service", nil)
	deliveryControlService := service.NewDeliveryControlService(
		&cfg.DeliveryControl,
		logger,
		repos.DeliveryControl,
		repos.Notification,
	)

	// Initialize notification service (without worker initially)
	logger.Info(logging.General, logging.Startup, "Initializing notification service", nil)
	notificationService := service.NewNotificationServiceWithDeliveryControl(
		cfg,
		logger,
		repos.Notification,
		repos.Template,
		repos.Preference,
		repos.Log,
		repos.Setting,
		repos.SMSTemplate,
		repos.Provider,
		repos.ProviderAttempt,
		nil, // Worker will be set after initialization
		wsHub,
		authClient,
		deliveryControlService,
	)

	// Initialize template service
	logger.Info(logging.General, logging.Startup, "Initializing template service", nil)
	templateService := service.NewTemplateService(repos.Template, logger)

	// Initialize preference service
	logger.Info(logging.General, logging.Startup, "Initializing preference service", nil)
	preferenceService := service.NewPreferenceService(repos.Preference, logger)

	// Initialize reminder service
	logger.Info(logging.General, logging.Startup, "Initializing reminder service", nil)
	reminderService := service.NewReminderService(
		repos.Reminder,
		notificationService,
		logger,
	)

	// Initialize provider balance service + scheduler (account balance/quota
	// monitoring and credit alerting)
	logger.Info(logging.General, logging.Startup, "Initializing provider balance service", nil)
	balanceService := service.NewBalanceService(
		cfg,
		logger,
		repos.Provider,
		repos.ProviderBalance,
	)
	balanceScheduler := service.NewBalanceScheduler(cfg, logger, balanceService)

	// Create handler adapters
	smsHandler := service.NewSMSHandlerAdapter(notificationService)
	emailHandler := service.NewEmailHandlerAdapter(notificationService)
	pushHandler := service.NewPushHandlerAdapter(notificationService)

	// Initialize digest service (without worker initially — SetWorker will be called after worker creation)
	logger.Info(logging.General, logging.Startup, "Initializing digest service", map[logging.ExtraKey]interface{}{
		"enabled":  cfg.Digest.Enabled,
		"interval": cfg.Digest.Interval,
	})
	digestService := service.NewDigestService(
		cfg,
		logger,
		repos.Notification,
		repos.Preference,
		smsHandler,
		emailHandler,
		nil, // Worker not yet created
	)

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
		repos.ProviderAttempt,
		smsHandler,
		emailHandler,
		pushHandler,
		digestService,
		deliveryControlService,
	)

	// Set worker on digest service (breaks circular dependency)
	digestService.SetWorker(notificationWorker)

	notificationWorker.Start()

	// Update service with worker reference
	notificationService = service.NewNotificationServiceWithDeliveryControl(
		cfg,
		logger,
		repos.Notification,
		repos.Template,
		repos.Preference,
		repos.Log,
		repos.Setting,
		repos.SMSTemplate,
		repos.Provider,
		repos.ProviderAttempt,
		notificationWorker,
		wsHub,
		authClient,
		deliveryControlService,
	)

	return &Services{
		Notification:      notificationService,
		Template:          templateService,
		Preference:        preferenceService,
		Reminder:          reminderService,
		Worker:            notificationWorker,
		WebSocketHub:      wsHub,
		AuthClient:        authClient,
		Balance:           balanceService,
		BalanceScheduler:  balanceScheduler,
		DeliveryControl:   deliveryControlService,
		RetentionScheduler: initRetention(repos.DB, logger),
	}
}

// initRetention creates the retention scheduler and its dependencies.
func initRetention(db *gorm.DB, logger logging.Logger) *retention.Scheduler {
	policyRepo := retention.NewPolicyRepository(db, logger)
	runRepo := retention.NewRunRepository(db, logger)
	runner := retention.NewNotifierRunner(db, logger)
	lock := retention.NewPGLock(db)
	scheduler := retention.NewScheduler(policyRepo, runRepo, runner, lock, logger)
	scheduler.Start()
	return scheduler
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
