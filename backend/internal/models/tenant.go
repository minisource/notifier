package models

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant/project/organization in the notifier system.
// Each tenant isolates notifications, templates, providers, reminders, etc.
//
// The TenantMiddleware (api/api.go) validates the X-Tenant-ID header against
// this table using `id = ? AND is_active = true`, so the IsActive field maps
// to the `is_active` column.
type Tenant struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Slug      string    `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	DisplayName string  `gorm:"size:255" json:"displayName,omitempty"`
	Description string  `gorm:"size:1000" json:"description,omitempty"`
	IsActive  bool      `gorm:"default:true" json:"isActive"`
	IsDefault bool      `gorm:"default:false" json:"isDefault"` // Default/global tenant

	// Enabled channels as JSON array, e.g. ["email","sms","push","in_app"]
	// serializer:json handles both jsonb DB storage and JSON array output
	EnabledChannels []string `gorm:"type:jsonb;default:'[]';serializer:json" json:"enabledChannels"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (t *Tenant) TableName() string {
	return "tenants"
}
