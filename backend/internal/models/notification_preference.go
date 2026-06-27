package models

import (
	"encoding/json"
	"fmt"
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

// SetQuietHours sets the QuietHours field from a config struct
func (np *NotificationPreference) SetQuietHours(qh *QuietHoursConfig) error {
	data, err := json.Marshal(qh)
	if err != nil {
		return err
	}
	np.QuietHours = string(data)
	return nil
}

// SetCategorySettings sets the CategorySettings field from a map
func (np *NotificationPreference) SetCategorySettings(settings map[NotificationCategory]bool) error {
	if settings == nil {
		np.CategorySettings = "{}"
		return nil
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	np.CategorySettings = string(data)
	return nil
}

// TableName specifies the table name
func (NotificationPreference) TableName() string {
	return "notification_preferences"
}

// NotificationCategory represents a category for notification types (used in preference filtering)
type NotificationCategory string

const (
	NotificationCategorySystem    NotificationCategory = "system"
	NotificationCategoryAlerts    NotificationCategory = "alerts"
	NotificationCategoryUpdates   NotificationCategory = "updates"
	NotificationCategoryMarketing NotificationCategory = "marketing"
	NotificationCategorySecurity  NotificationCategory = "security"
)

// QuietHoursConfig represents the parsed quiet hours configuration
type QuietHoursConfig struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	Timezone string `json:"timezone"`
}

// ParseQuietHours parses the QuietHours JSON string into a struct
func (np *NotificationPreference) ParseQuietHours() (*QuietHoursConfig, error) {
	if np.QuietHours == "" || np.QuietHours == "{}" {
		return nil, nil
	}
	var qh QuietHoursConfig
	if err := json.Unmarshal([]byte(np.QuietHours), &qh); err != nil {
		return nil, err
	}
	return &qh, nil
}

// IsInQuietHours checks if the current time is within the user's quiet hours
func (np *NotificationPreference) IsInQuietHours() bool {
	qh, err := np.ParseQuietHours()
	if err != nil || qh == nil || qh.Start == "" || qh.End == "" {
		return false
	}

	now := time.Now()
	currentMinutes := now.Hour()*60 + now.Minute()

	startMinutes := parseTimeToMinutes(qh.Start)
	endMinutes := parseTimeToMinutes(qh.End)

	if startMinutes < 0 || endMinutes < 0 {
		return false
	}

	// Handle overnight quiet hours (e.g., 22:00 - 08:00)
	if startMinutes <= endMinutes {
		// Same day (e.g., 09:00 - 17:00)
		return currentMinutes >= startMinutes && currentMinutes <= endMinutes
	}
	// Overnight (e.g., 22:00 - 08:00)
	return currentMinutes >= startMinutes || currentMinutes <= endMinutes
}

// parseTimeToMinutes parses a time string (HH:MM) into minutes since midnight
func parseTimeToMinutes(timeStr string) int {
	if len(timeStr) < 5 {
		return -1
	}
	hours := 0
	minutes := 0
	if _, err := fmt.Sscanf(timeStr, "%d:%d", &hours, &minutes); err != nil {
		return -1
	}
	if hours < 0 || hours > 23 || minutes < 0 || minutes > 59 {
		return -1
	}
	return hours*60 + minutes
}

// ParseCategorySettings parses the CategorySettings JSON string into a map
func (np *NotificationPreference) ParseCategorySettings() map[NotificationCategory]bool {
	if np.CategorySettings == "" || np.CategorySettings == "{}" {
		return nil
	}
	var result map[NotificationCategory]bool
	if err := json.Unmarshal([]byte(np.CategorySettings), &result); err != nil {
		return nil
	}
	return result
}

// IsCategoryEnabled checks if a specific notification category is enabled
func (np *NotificationPreference) IsCategoryEnabled(category NotificationCategory) bool {
	cats := np.ParseCategorySettings()
	if cats == nil {
		return true // All categories enabled by default
	}
	enabled, ok := cats[category]
	if !ok {
		return true // Unknown category defaults to enabled
	}
	return enabled
}
