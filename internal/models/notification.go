package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeSMS   NotificationType = "sms"
	NotificationTypeEmail NotificationType = "email"
	NotificationTypePush  NotificationType = "push"
	NotificationTypeInApp NotificationType = "in_app"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending  NotificationStatus = "pending"
	NotificationStatusSending  NotificationStatus = "sending"
	NotificationStatusSent     NotificationStatus = "sent"
	NotificationStatusFailed   NotificationStatus = "failed"
	NotificationStatusRetrying NotificationStatus = "retrying"
	NotificationStatusCanceled NotificationStatus = "canceled"
)

// NotificationPriority represents the priority level
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

// Notification represents a notification record in the database
type Notification struct {
	ID       uuid.UUID            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID *uuid.UUID           `gorm:"type:uuid;index" json:"tenantId,omitempty"`
	UserID   uuid.UUID            `gorm:"type:uuid;not null;index" json:"userId"`
	Type     NotificationType     `gorm:"type:varchar(20);not null;index" json:"type"`
	Status   NotificationStatus   `gorm:"type:varchar(20);not null;index;default:'pending'" json:"status"`
	Priority NotificationPriority `gorm:"type:varchar(20);not null;default:'normal'" json:"priority"`

	// Recipient information
	RecipientEmail string `gorm:"type:varchar(255);index" json:"recipientEmail,omitempty"`
	RecipientPhone string `gorm:"type:varchar(20);index" json:"recipientPhone,omitempty"`
	RecipientID    string `gorm:"type:varchar(255);index" json:"recipientId,omitempty"` // For push notifications

	// Content
	Subject  string `gorm:"type:varchar(500)" json:"subject,omitempty"`
	Body     string `gorm:"type:text;not null" json:"body"`
	Metadata string `gorm:"type:jsonb" json:"metadata,omitempty"` // Additional data as JSON

	// Template information
	TemplateID *uuid.UUID            `gorm:"type:uuid;index" json:"templateId,omitempty"`
	Template   *NotificationTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`

	// Retry information
	RetryCount  int        `gorm:"default:0" json:"retryCount"`
	MaxRetries  int        `gorm:"default:3" json:"maxRetries"`
	NextRetryAt *time.Time `gorm:"index" json:"nextRetryAt,omitempty"`

	// Error information
	ErrorMessage string `gorm:"type:text" json:"errorMessage,omitempty"`

	// Provider information
	Provider      string `gorm:"type:varchar(100)" json:"provider,omitempty"`
	ProviderMsgID string `gorm:"type:varchar(255);index" json:"providerMsgId,omitempty"`

	// Timing
	ScheduledAt *time.Time `gorm:"index" json:"scheduledAt,omitempty"`
	SentAt      *time.Time `gorm:"index" json:"sentAt,omitempty"`
	ReadAt      *time.Time `json:"readAt,omitempty"`

	// Audit fields
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Notification) TableName() string {
	return "notifications"
}
