package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ProviderStatusActive   = "active"
	ProviderStatusInactive = "inactive"
	ProviderStatusDisabled = "disabled"
	ProviderStatusError    = "error"
)

// Provider represents a notification provider (channel) configuration
type Provider struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     *uuid.UUID     `gorm:"type:uuid;index" json:"tenantId,omitempty"`
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	Channel      string         `gorm:"type:varchar(50);not null;index" json:"channel"` // sms, email, push, in_app, webhook
	Type         string         `gorm:"type:varchar(100)" json:"type,omitempty"`         // e.g., kavenegar, smtp, fcm
	Status       string         `gorm:"type:varchar(20);not null;default:active" json:"status"`
	Config       string         `gorm:"type:text" json:"config,omitempty"`               // JSON config (secrets are redacted in responses)
	SecretConfig string         `gorm:"type:text" json:"secretConfig,omitempty"`         // JSON encrypted secrets (never exposed in API responses)
	Priority     int            `gorm:"not null;default:1" json:"priority"`
	IsEnabled    bool           `gorm:"not null;default:true" json:"isEnabled"`
	IsPrimary    bool           `gorm:"not null;default:false" json:"isPrimary"`
	IsDefault    bool           `gorm:"not null;default:false" json:"isDefault"`
	Description  string         `gorm:"type:text" json:"description,omitempty"`
	CreatedAt    time.Time      `gorm:"not null;default:now()" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"not null;default:now()" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (p *Provider) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Status == "" {
		p.Status = ProviderStatusActive
	}
	return nil
}

// TableName specifies the table name
func (Provider) TableName() string {
	return "providers"
}
