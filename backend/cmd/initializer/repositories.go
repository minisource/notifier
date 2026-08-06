package initializer

import (
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/repository"
	"gorm.io/gorm"
)

// Repositories holds all repository instances
type Repositories struct {
	DB          *gorm.DB
	Notification repository.NotificationRepository
	Template     repository.NotificationTemplateRepository
	Preference   repository.NotificationPreferenceRepository
	Log          repository.NotificationLogRepository
	Setting      repository.SettingRepository
	SMSTemplate  repository.SMSTemplateRepository
	Reminder     repository.ReminderRepository
	Provider     repository.ProviderRepository
	Tenant       repository.TenantRepository
	ProviderAttempt repository.ProviderAttemptRepository
	ProviderBalance repository.ProviderBalanceRepository
	DeliveryControl repository.DeliveryControlRepository
}

// InitRepositories creates all repository instances
func InitRepositories(db *gorm.DB, logger logging.Logger) *Repositories {
	return &Repositories{
		DB:         db,
		Notification: repository.NewNotificationRepository(db, logger),
		Template:     repository.NewNotificationTemplateRepository(db, logger),
		Preference:   repository.NewNotificationPreferenceRepository(db, logger),
		Log:          repository.NewNotificationLogRepository(db, logger),
		Setting:      repository.NewSettingRepository(db, logger),
		SMSTemplate:  repository.NewSMSTemplateRepository(db, logger),
		Reminder:     repository.NewReminderRepository(db, logger),
		Provider:     repository.NewProviderRepository(db, logger),
		Tenant:       repository.NewTenantRepository(db, logger),
		ProviderAttempt: repository.NewProviderAttemptRepository(db, logger),
		ProviderBalance: repository.NewProviderBalanceRepository(db, logger),
		DeliveryControl: repository.NewDeliveryControlRepository(db, logger),
	}
}
