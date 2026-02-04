package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Setting represents application settings stored in database
type Setting struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    *uuid.UUID     `gorm:"type:uuid;index;uniqueIndex:idx_setting_key_tenant,priority:1" json:"tenantId,omitempty"`
	Key         string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_setting_key_tenant,priority:2" json:"key"`
	Value       string         `gorm:"type:text;not null" json:"value"`
	Category    string         `gorm:"type:varchar(100);not null;index" json:"category"` // sms, email, notification, system
	Description string         `gorm:"type:text" json:"description,omitempty"`
	IsEncrypted bool           `gorm:"not null;default:false" json:"isEncrypted"`
	IsActive    bool           `gorm:"not null;default:true" json:"isActive"`
	CreatedAt   time.Time      `gorm:"not null;default:now()" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (s *Setting) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Setting) TableName() string {
	return "settings"
}

// Common setting keys
const (
	SettingKeySMSProviders         = "sms.providers"
	SettingKeySMSDefaultProvider   = "sms.default_provider"
	SettingKeyEmailProviders       = "email.providers"
	SettingKeyEmailDefaultProvider = "email.default_provider"
	SettingKeyNotificationRetries  = "notification.max_retries"
	SettingKeyNotificationTimeout  = "notification.timeout_seconds"
	SettingKeyWorkerPoolSize       = "worker.pool_size"
	SettingKeyWorkerQueueSize      = "worker.queue_size"
)
