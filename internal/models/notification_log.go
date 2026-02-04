package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationLog represents detailed logs of notification operations
type NotificationLog struct {
	ID             uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID       *uuid.UUID    `gorm:"type:uuid;index" json:"tenantId,omitempty"`
	NotificationID uuid.UUID     `gorm:"type:uuid;not null;index" json:"notificationId"`
	Notification   *Notification `gorm:"foreignKey:NotificationID" json:"notification,omitempty"`

	// Log details
	Action       string             `gorm:"type:varchar(50);not null" json:"action"` // created, sending, sent, failed, retrying
	Status       NotificationStatus `gorm:"type:varchar(20);not null" json:"status"`
	Message      string             `gorm:"type:text" json:"message,omitempty"`
	ErrorDetails string             `gorm:"type:text" json:"errorDetails,omitempty"`

	// Provider response
	ProviderResponse string `gorm:"type:jsonb" json:"providerResponse,omitempty"`

	// Performance metrics
	ProcessingTimeMs int `gorm:"default:0" json:"processingTimeMs"`

	// Audit fields
	CreatedAt time.Time `gorm:"not null;default:now();index" json:"createdAt"`
}

// BeforeCreate hook to generate UUID and set default JSON values
func (nl *NotificationLog) BeforeCreate(tx *gorm.DB) error {
	if nl.ID == uuid.Nil {
		nl.ID = uuid.New()
	}
	// Set default JSON values for jsonb fields to avoid PostgreSQL validation errors
	if nl.ProviderResponse == "" {
		nl.ProviderResponse = "{}"
	}
	return nil
}

// TableName specifies the table name
func (NotificationLog) TableName() string {
	return "notification_logs"
}
