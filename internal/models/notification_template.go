package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationTemplate represents a reusable notification template
type NotificationTemplate struct {
	ID          uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    *uuid.UUID       `gorm:"type:uuid;index" json:"tenantId,omitempty"`
	Name        string           `gorm:"type:varchar(255);not null;uniqueIndex:idx_template_name_type_tenant,priority:2" json:"name"`
	Type        NotificationType `gorm:"type:varchar(20);not null;uniqueIndex:idx_template_name_type_tenant,priority:3" json:"type"`
	Subject     string           `gorm:"type:varchar(500)" json:"subject,omitempty"`
	Body        string           `gorm:"type:text;not null" json:"body"`
	Description string           `gorm:"type:text" json:"description,omitempty"`

	// Template variables (JSON array of variable names)
	Variables string `gorm:"type:jsonb" json:"variables,omitempty"` // e.g., ["userName", "code", "expiryTime"]

	// Provider specific settings
	Provider         string `gorm:"type:varchar(100)" json:"provider,omitempty"`
	ProviderTemplate string `gorm:"type:varchar(255)" json:"providerTemplate,omitempty"`

	// Status
	IsActive bool `gorm:"not null;default:true" json:"isActive"`

	// Audit fields
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (nt *NotificationTemplate) BeforeCreate(tx *gorm.DB) error {
	if nt.ID == uuid.Nil {
		nt.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (NotificationTemplate) TableName() string {
	return "notification_templates"
}
