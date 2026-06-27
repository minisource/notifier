package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReminderStatus represents the lifecycle state of a scheduled reminder
type ReminderStatus string

const (
	ReminderStatusPending    ReminderStatus = "pending"    // Scheduled, waiting to fire
	ReminderStatusProcessing ReminderStatus = "processing" // Currently being processed
	ReminderStatusSent       ReminderStatus = "sent"       // Reminder fired, notification created
	ReminderStatusFailed     ReminderStatus = "failed"     // Failed to create notification
	ReminderStatusCancelled  ReminderStatus = "cancelled"  // Manually cancelled before firing
)

// Reminder represents a scheduled notification that fires at a specific time
type Reminder struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  *uuid.UUID     `gorm:"type:uuid;index" json:"tenantId,omitempty"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId"`
	ProjectID string         `gorm:"type:varchar(100);index" json:"projectId,omitempty"`

	// Scheduling
	ScheduledAt time.Time      `gorm:"not null;index:idx_reminders_due,priority:1" json:"scheduledAt"`
	Status      ReminderStatus `gorm:"type:varchar(20);not null;default:'pending';index:idx_reminders_due,priority:2" json:"status"`

	// Recipient
	RecipientEmail string           `gorm:"type:varchar(255)" json:"recipientEmail,omitempty"`
	RecipientPhone string           `gorm:"type:varchar(20)" json:"recipientPhone,omitempty"`
	Channels       []NotificationType `gorm:"-" json:"channels,omitempty"` // Virtual field for request parsing
	ChannelsJSON   string           `gorm:"column:channels;type:jsonb;not null;default:'[]'" json:"-"`

	// Content — template key + variables
	TemplateKey string `gorm:"type:varchar(255)" json:"templateKey,omitempty"`
	Subject     string `gorm:"type:varchar(500)" json:"subject,omitempty"`
	Body        string `gorm:"type:text" json:"body,omitempty"`

	// Template variables as JSON
	VariablesJSON string `gorm:"column:variables;type:jsonb;default:'{}'" json:"-"`

	// Result
	NotificationID *uuid.UUID `gorm:"type:uuid;index" json:"notificationId,omitempty"` // Notification created when reminder fires
	SentAt         *time.Time `json:"sentAt,omitempty"`
	CancelledAt    *time.Time `json:"cancelledAt,omitempty"`
	LastError      string     `gorm:"type:text" json:"lastError,omitempty"`

	// Audit
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (r *Reminder) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Reminder) TableName() string {
	return "reminders"
}

// ParseChannels deserializes ChannelsJSON into the Channels field
func (r *Reminder) ParseChannels() []NotificationType {
	if r.ChannelsJSON == "" || r.ChannelsJSON == "[]" {
		return nil
	}
	var chs []NotificationType
	if err := json.Unmarshal([]byte(r.ChannelsJSON), &chs); err != nil {
		return nil
	}
	return chs
}

// SetChannels serializes Channels into ChannelsJSON
func (r *Reminder) SetChannels(chs []NotificationType) error {
	data, err := json.Marshal(chs)
	if err != nil {
		return err
	}
	r.ChannelsJSON = string(data)
	return nil
}

// ParseVariables deserializes the VariablesJSON into a map
func (r *Reminder) ParseVariables() map[string]string {
	if r.VariablesJSON == "" || r.VariablesJSON == "{}" {
		return nil
	}
	var vars map[string]string
	if err := json.Unmarshal([]byte(r.VariablesJSON), &vars); err != nil {
		return nil
	}
	return vars
}

// SetVariables serializes a variables map into VariablesJSON
func (r *Reminder) SetVariables(vars map[string]string) error {
	data, err := json.Marshal(vars)
	if err != nil {
		return err
	}
	r.VariablesJSON = string(data)
	return nil
}
