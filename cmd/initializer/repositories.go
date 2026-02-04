package initializer

import (
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/repository"
	"gorm.io/gorm"
)

// Repositories holds all repository instances
type Repositories struct {
	Notification repository.NotificationRepository
	Template     repository.NotificationTemplateRepository
	Preference   repository.NotificationPreferenceRepository
	Log          repository.NotificationLogRepository
	Setting      repository.SettingRepository
	SMSTemplate  repository.SMSTemplateRepository
}

// InitRepositories creates all repository instances
func InitRepositories(db *gorm.DB, logger logging.Logger) *Repositories {
	return &Repositories{
		Notification: repository.NewNotificationRepository(db, logger),
		Template:     repository.NewNotificationTemplateRepository(db, logger),
		Preference:   repository.NewNotificationPreferenceRepository(db, logger),
		Log:          repository.NewNotificationLogRepository(db, logger),
		Setting:      repository.NewSettingRepository(db, logger),
		SMSTemplate:  repository.NewSMSTemplateRepository(db, logger),
	}
}
