package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationPreference represents user's notification preferences
type NotificationPreference struct {
	ID       uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID *uuid.UUID       `gorm:"type:uuid;index" json:"tenantId,omitempty"`
	UserID   uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_user_type_tenant,priority:2" json:"userId"`
	Type     NotificationType `gorm:"type:varchar(20);not null;uniqueIndex:idx_user_type_tenant,priority:3" json:"type"`

	// Channel enablement
	IsEnabled bool `gorm:"not null;default:true" json:"isEnabled"`

	// Frequency settings
	AllowInstant    bool   `gorm:"not null;default:true" json:"allowInstant"`               // Receive notifications immediately
	AllowDigest     bool   `gorm:"not null;default:false" json:"allowDigest"`               // Receive daily/weekly digest
	DigestFrequency string `gorm:"type:varchar(20);default:'daily'" json:"digestFrequency"` // daily, weekly, monthly

	// Quiet hours (stored as JSON with start and end times)
	QuietHours string `gorm:"type:jsonb" json:"quietHours,omitempty"` // e.g., {"start": "22:00", "end": "08:00", "timezone": "UTC"}

	// Category preferences (JSON object with category settings)
	CategorySettings string `gorm:"type:jsonb" json:"categorySettings,omitempty"` // e.g., {"marketing": false, "alerts": true, "updates": true}

	// Audit fields
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (np *NotificationPreference) BeforeCreate(tx *gorm.DB) error {
	if np.ID == uuid.Nil {
		np.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (NotificationPreference) TableName() string {
	return "notification_preferences"
}
