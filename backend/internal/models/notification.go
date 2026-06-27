package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType represents the type (channel) of notification
type NotificationType string

const (
	NotificationTypeSMS      NotificationType = "sms"
	NotificationTypeEmail    NotificationType = "email"
	NotificationTypePush     NotificationType = "push"
	NotificationTypeInApp    NotificationType = "in_app"
	NotificationTypeWebhook  NotificationType = "webhook"
	NotificationTypeSecurity NotificationType = "security"
)

// NotificationChannel is an alias for NotificationType for semantic clarity in DTOs
type NotificationChannel = NotificationType

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending    NotificationStatus = "pending"
	NotificationStatusQueued     NotificationStatus = "queued"
	NotificationStatusSending    NotificationStatus = "sending"
	NotificationStatusProcessing NotificationStatus = "processing"
	NotificationStatusSent       NotificationStatus = "sent"
	NotificationStatusDelivered  NotificationStatus = "delivered"
	NotificationStatusFailed     NotificationStatus = "failed"
	NotificationStatusRetrying   NotificationStatus = "retrying"
	NotificationStatusDead       NotificationStatus = "dead"
	NotificationStatusCanceled   NotificationStatus = "canceled"
	NotificationStatusCancelled  NotificationStatus = "cancelled"
	NotificationStatusDigested   NotificationStatus = "digested"
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
	TemplateID  *uuid.UUID            `gorm:"type:uuid;index" json:"templateId,omitempty"`
	TemplateKey string                `gorm:"type:varchar(255);index" json:"templateKey,omitempty"`
	Template    *NotificationTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`

	// Retry information
	RetryCount  int        `gorm:"default:0" json:"retryCount"`
	MaxRetries  int        `gorm:"default:3" json:"maxRetries"`
	NextRetryAt *time.Time `gorm:"index" json:"nextRetryAt,omitempty"`

	// Error information
	ErrorMessage string `gorm:"type:text" json:"errorMessage,omitempty"`

	// Locale — user's preferred language for this notification (e.g., "en", "fa", "ar")
	Locale string `gorm:"type:varchar(10);not null;default:'en';index" json:"locale"`

	// Idempotency key — prevents duplicate sends for the same logical notification
	IdempotencyKey string `gorm:"type:varchar(255);uniqueIndex:idx_notif_idempotency_key;default:''" json:"idempotencyKey,omitempty"`

	// Provider information
	Provider      string `gorm:"type:varchar(100)" json:"provider,omitempty"`
	ProviderMsgID string `gorm:"type:varchar(255);index" json:"providerMsgId,omitempty"`

	// Timing
	ScheduledAt  *time.Time `gorm:"index" json:"scheduledAt,omitempty"`
	SentAt       *time.Time `gorm:"index" json:"sentAt,omitempty"`
	DeliveredAt  *time.Time `json:"deliveredAt,omitempty"`
	FailedAt     *time.Time `json:"failedAt,omitempty"`
	SeenAt       *time.Time `json:"seenAt,omitempty"`
	ReadAt       *time.Time `json:"readAt,omitempty"`
	ClickedAt    *time.Time `json:"clickedAt,omitempty"`
	CancelledAt  *time.Time `json:"cancelledAt,omitempty"`

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
