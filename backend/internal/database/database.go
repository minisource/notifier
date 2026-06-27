package database

import (
	"fmt"
	"os"
	"time"

	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// InitDatabase initializes the database connection and runs migrations
func InitDatabase(cfg *config.PostgresConfig, dbCfg *config.DatabaseConfig, logger logging.Logger) (*gorm.DB, error) {
	logger.Info(logging.Postgres, logging.Startup, "Initializing database connection", map[logging.ExtraKey]interface{}{
		"host":   cfg.Host,
		"port":   cfg.Port,
		"dbName": cfg.DbName,
	})

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DbName,
		cfg.SSLMode,
	)

	// Configure GORM logger
	gormLogLevel := gormLogger.Silent
	if cfg.SSLMode == "disable" { // Development mode
		gormLogLevel = gormLogger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		logger.Error(logging.Postgres, logging.Startup, "Failed to connect to database", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error(logging.Postgres, logging.Startup, "Failed to get SQL DB instance", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	logger.Debug(logging.Postgres, logging.Startup, "Database connection pool configured", map[logging.ExtraKey]interface{}{
		"maxIdleConns":    cfg.MaxIdleConns,
		"maxOpenConns":    cfg.MaxOpenConns,
		"connMaxLifetime": cfg.ConnMaxLifetime,
	})

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		logger.Error(logging.Postgres, logging.Startup, "Failed to ping database", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	logger.Info(logging.Postgres, logging.Startup, "Database connection established successfully", nil)

	// Run migrations if enabled
	// SQL migrations are the source of truth for production schema.
	// GORM AutoMigrate is only for local development when explicitly enabled.
	if dbCfg.RunMigrations {
		if err := RunMigrations(db, dbCfg.RunSeedData, dbCfg.AutoMigrate, logger); err != nil {
			logger.Error(logging.Postgres, logging.Startup, "Failed to run migrations", map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
			return nil, err
		}
	} else {
		logger.Info(logging.Postgres, logging.Startup, "Database migrations skipped (disabled in config)", nil)
	}

	return db, nil
}

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB, runSeedData bool, autoMigrate bool, logger logging.Logger) error {
	logger.Info(logging.Postgres, logging.Startup, "Running database migrations", nil)

	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		logger.Error(logging.Postgres, logging.Startup, "Failed to create UUID extension", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return err
	}

	logger.Debug(logging.Postgres, logging.Startup, "UUID extension enabled", nil)

	// GORM AutoMigrate — only for local development when explicitly enabled
	// SQL migrations (via golang-migrate) are the source of truth for production.
	if autoMigrate {
		logger.Info(logging.Postgres, logging.Startup, "GORM AutoMigrate is ENABLED (development mode)", nil)

		models := []interface{}{
			&models.NotificationTemplate{},
			&models.NotificationPreference{},
			&models.Notification{},
			&models.NotificationLog{},
			&models.Setting{},
			&models.ServiceClient{},
			&models.SMSTemplate{},
			&models.Provider{},
			&models.Reminder{},
		}

		for _, model := range models {
			logger.Debug(logging.Postgres, logging.Startup, "Migrating model", map[logging.ExtraKey]interface{}{
				"model": fmt.Sprintf("%T", model),
			})

			if err := db.AutoMigrate(model); err != nil {
				logger.Error(logging.Postgres, logging.Startup, "Failed to migrate model", map[logging.ExtraKey]interface{}{
					"model": fmt.Sprintf("%T", model),
					"error": err.Error(),
				})
				return err
			}
		}

		logger.Info(logging.Postgres, logging.Startup, "GORM AutoMigrate completed", nil)
	} else {
		logger.Info(logging.Postgres, logging.Startup, "GORM AutoMigrate is DISABLED (production mode — use SQL migrations)", nil)
	}

	logger.Info(logging.Postgres, logging.Startup, "Database migrations completed successfully", nil)

	// Create indexes for better performance
	if err := createIndexes(db, logger); err != nil {
		logger.Warn(logging.Postgres, logging.Startup, "Failed to create some indexes", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
	}

	// Seed default data if enabled
	if runSeedData {
		if err := seedDefaultData(db, logger); err != nil {
			logger.Warn(logging.Postgres, logging.Startup, "Failed to seed default data", map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
		}
	} else {
		logger.Info(logging.Postgres, logging.Startup, "Database seeding skipped (disabled in config)", nil)
	}

	return nil
}

// createIndexes creates additional indexes for better query performance
func createIndexes(db *gorm.DB, logger logging.Logger) error {
	logger.Debug(logging.Postgres, logging.Startup, "Creating additional indexes", nil)

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_notifications_status_priority ON notifications(status, priority DESC, created_at ASC)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_scheduled ON notifications(scheduled_at) WHERE scheduled_at IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_notifications_retry ON notifications(status, retry_count, next_retry_at) WHERE status = 'retrying'",
		"CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, read_at) WHERE read_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_notification_logs_created ON notification_logs(notification_id, created_at DESC)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			logger.Warn(logging.Postgres, logging.Startup, "Failed to create index", map[logging.ExtraKey]interface{}{
				"sql":   indexSQL,
				"error": err.Error(),
			})
			// Continue with other indexes even if one fails
		}
	}

	logger.Debug(logging.Postgres, logging.Startup, "Additional indexes created", nil)
	return nil
}

// seedDefaultData seeds default templates and settings
func seedDefaultData(db *gorm.DB, logger logging.Logger) error {
	logger.Info(logging.Postgres, logging.Startup, "Seeding default data", nil)

	// Seed default notification templates
	templates := []models.NotificationTemplate{
		{
			Name:        "welcome_email",
			Type:        "email",
			Subject:     "Welcome to {{.AppName}}",
			Body:        "Hello {{.Username}},\n\nWelcome to our platform!",
			Description: "Welcome email for new users",
			Variables:   "[]",
			IsActive:    true,
		},
		{
			Name:        "otp_email",
			Type:        "email",
			Subject:     "Your OTP Code",
			Body:        "Your OTP code is: {{.Code}}. It will expire in {{.Expiry}} minutes.",
			Description: "OTP verification email",
			Variables:   "[]",
			IsActive:    true,
		},
		{
			Name:        "password_reset",
			Type:        "email",
			Subject:     "Password Reset Request",
			Body:        "Your password reset code is: {{.Code}}",
			Description: "Password reset email",
			Variables:   "[]",
			IsActive:    true,
		},
	}

	for _, template := range templates {
		var existing models.NotificationTemplate
		if err := db.Where("name = ?", template.Name).First(&existing).Error; err != nil {
			if err := db.Create(&template).Error; err != nil {
				logger.Debug(logging.Postgres, logging.Startup, "Template already exists", map[logging.ExtraKey]interface{}{
					"template": template.Name,
				})
			} else {
				logger.Info(logging.Postgres, logging.Startup, "Created template", map[logging.ExtraKey]interface{}{
					"template": template.Name,
				})
			}
		}
	}

	// Seed service clients for service-to-service authentication
	authClientID := os.Getenv("AUTH_SERVICE_CLIENT_ID")
	authClientSecret := os.Getenv("AUTH_SERVICE_CLIENT_SECRET")

	if authClientID == "" {
		authClientID = "auth-service"
	}
	if authClientSecret == "" {
		authClientSecret = "auth-service-secret-key"
	}

	serviceClients := []models.ServiceClient{
		{
			Name:         "Auth Service",
			ClientID:     authClientID,
			ClientSecret: authClientSecret, // In production, this should be hashed
			Description:  "Authentication service client for sending OTP and notifications",
			Scopes:       "notifications:send,notifications:create,templates:read",
			IsActive:     true,
		},
	}

	for _, client := range serviceClients {
		var existing models.ServiceClient
		if err := db.Where("client_id = ?", client.ClientID).First(&existing).Error; err != nil {
			if err := db.Create(&client).Error; err != nil {
				logger.Debug(logging.Postgres, logging.Startup, "Service client already exists", map[logging.ExtraKey]interface{}{
					"client_id": client.ClientID,
				})
			} else {
				logger.Info(logging.Postgres, logging.Startup, "Created service client", map[logging.ExtraKey]interface{}{
					"client_id": client.ClientID,
					"name":      client.Name,
				})
			}
		}
	}

	// Seed SMS provider configuration
	// Use Kavenegar if configured, otherwise use mock provider
	cfg := config.GetConfig()
	var smsProviderConfig string
	var smsProviderDesc string

	if cfg.Kavenegar.Enabled && cfg.Kavenegar.APIKey != "" {
		// Use Kavenegar provider
		smsProviderConfig = fmt.Sprintf(`{"provider":"kavenegar","apiKey":"%s","template":"%s"}`,
			cfg.Kavenegar.APIKey, cfg.Kavenegar.Template)
		smsProviderDesc = "SMS provider configuration (Kavenegar)"
		logger.Info(logging.Postgres, logging.Startup, "Using Kavenegar SMS provider", map[logging.ExtraKey]interface{}{
			"template": cfg.Kavenegar.Template,
		})
	} else {
		// Use mock provider for development
		smsProviderConfig = `{"provider":"mock","apiKey":"dev-mock-key","template":"Your OTP code is: %s"}`
		smsProviderDesc = "SMS provider configuration (mock for development)"
		logger.Info(logging.Postgres, logging.Startup, "Using mock SMS provider (set KAVENEGAR_ENABLED=true to use Kavenegar)", nil)
	}

	smsSettings := []models.Setting{
		{
			Key:         models.SettingKeySMSProviders,
			Value:       smsProviderConfig,
			Category:    "sms",
			Description: smsProviderDesc,
			IsActive:    true,
			IsEncrypted: false,
		},
	}

	for _, setting := range smsSettings {
		var existing models.Setting
		if err := db.Where("key = ?", setting.Key).First(&existing).Error; err != nil {
			if err := db.Create(&setting).Error; err != nil {
				logger.Debug(logging.Postgres, logging.Startup, "SMS setting already exists", map[logging.ExtraKey]interface{}{
					"key": setting.Key,
				})
			} else {
				logger.Info(logging.Postgres, logging.Startup, "Created SMS setting", map[logging.ExtraKey]interface{}{
					"key":      setting.Key,
					"provider": cfg.Kavenegar.Enabled,
				})
			}
		} else {
			// Update existing setting if provider changed
			if existing.Value != smsProviderConfig {
				existing.Value = smsProviderConfig
				existing.Description = smsProviderDesc
				if err := db.Save(&existing).Error; err != nil {
					logger.Warn(logging.Postgres, logging.Startup, "Failed to update SMS setting", map[logging.ExtraKey]interface{}{
						"key":   setting.Key,
						"error": err.Error(),
					})
				} else {
					logger.Info(logging.Postgres, logging.Startup, "Updated SMS setting", map[logging.ExtraKey]interface{}{
						"key": setting.Key,
					})
				}
			}
		}
	}

	// Seed default providers for development/testing
	defaultProviders := []models.Provider{
		{
			Name:        "Mock SMS",
			Channel:     "sms",
			Type:        "mock-sms",
			Status:      models.ProviderStatusActive,
			Config:      `{"sender":"10008663","baseUrl":"https://api.mock-sms.local"}`,
			SecretConfig: `{"apiKey":"mock-sms-dev-key-12345"}`,
			Priority:    10,
			IsEnabled:   true,
			IsDefault:   true,
			IsPrimary:   true,
			Description: "Mock SMS provider for development and testing",
		},
		{
			Name:        "Mock Email",
			Channel:     "email",
			Type:        "mock-email",
			Status:      models.ProviderStatusActive,
			Config:      `{"fromAddress":"noreply@mock-email.local","fromName":"Notifier Dev"}`,
			SecretConfig: `{"apiKey":"mock-email-dev-key-12345"}`,
			Priority:    10,
			IsEnabled:   true,
			IsDefault:   true,
			IsPrimary:   true,
			Description: "Mock Email provider for development and testing",
		},
	}

	for _, provider := range defaultProviders {
		var existing models.Provider
		if err := db.Where("name = ?", provider.Name).First(&existing).Error; err != nil {
			if err := db.Create(&provider).Error; err != nil {
				logger.Debug(logging.Postgres, logging.Startup, "Provider already exists", map[logging.ExtraKey]interface{}{
					"name": provider.Name,
				})
			} else {
				logger.Info(logging.Postgres, logging.Startup, "Created default provider", map[logging.ExtraKey]interface{}{
					"name":    provider.Name,
					"channel": provider.Channel,
				})
			}
		} else {
			logger.Debug(logging.Postgres, logging.Startup, "Provider already exists, skipping", map[logging.ExtraKey]interface{}{
				"name": provider.Name,
			})
		}
	}

	// Seed SMS templates for Kavenegar provider
	smsTemplates := []models.SMSTemplate{
		{
			Key:              "verify", // Match auth service template name
			Provider:         "kavenegar",
			ProviderTemplate: "verify",
			Description:      "OTP verification code template",
		},
		{
			Key:              "welcome",
			Provider:         "kavenegar",
			ProviderTemplate: "welcome",
			Description:      "Welcome message template",
		},
		{
			Key:              "order_placed",
			Provider:         "kavenegar",
			ProviderTemplate: "orderplaced",
			Description:      "Order confirmation template",
		},
		{
			Key:              "payment_success",
			Provider:         "kavenegar",
			ProviderTemplate: "paymentsuccess",
			Description:      "Payment success notification",
		},
	}

	// Set token mappings
	tokenMappings := []struct {
		key      string
		tokenMap map[string]string
	}{
		{"verify", map[string]string{"code": "token"}}, // Match auth service template name
		{"welcome", map[string]string{"name": "token"}},
		{"order_placed", map[string]string{"order_id": "token", "amount": "token2"}},
		{"payment_success", map[string]string{"amount": "token", "transaction_id": "token2"}},
	}

	for _, mapping := range tokenMappings {
		for i := range smsTemplates {
			if smsTemplates[i].Key == mapping.key {
				smsTemplates[i].SetTokenMapping(mapping.tokenMap)
				break
			}
		}
	}

	// Insert SMS templates
	for _, template := range smsTemplates {
		var existing models.SMSTemplate
		if err := db.Where("key = ? AND provider = ?", template.Key, template.Provider).First(&existing).Error; err != nil {
			if err := db.Create(&template).Error; err != nil {
				logger.Debug(logging.Postgres, logging.Startup, "SMS template already exists", map[logging.ExtraKey]interface{}{
					"key": template.Key,
				})
			} else {
				logger.Info(logging.Postgres, logging.Startup, "Created SMS template", map[logging.ExtraKey]interface{}{
					"key":      template.Key,
					"provider": template.Provider,
				})
			}
		}
	}

	logger.Info(logging.Postgres, logging.Startup, "Default data seeded successfully", nil)
	return nil
}

// CloseDatabase closes the database connection
func CloseDatabase(db *gorm.DB, logger logging.Logger) error {
	logger.Info(logging.Postgres, logging.Startup, "Closing database connection", nil)

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error(logging.Postgres, logging.Startup, "Failed to get SQL DB instance", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return err
	}

	if err := sqlDB.Close(); err != nil {
		logger.Error(logging.Postgres, logging.Startup, "Failed to close database connection", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return err
	}

	logger.Info(logging.Postgres, logging.Startup, "Database connection closed successfully", nil)
	return nil
}
