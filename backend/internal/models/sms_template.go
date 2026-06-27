package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SMSTemplate represents a provider-specific SMS template mapping
// For lookup-based providers (Kavenegar): uses ProviderTemplate
// For text-based providers (Twilio, etc.): uses MessageTemplate with {{placeholders}}
type SMSTemplate struct {
	ID       uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID *uuid.UUID `gorm:"type:uuid;index" json:"tenantId,omitempty"`

	// Template identification
	Key      string `gorm:"type:varchar(100);not null;uniqueIndex:idx_sms_templates_key_provider_tenant,priority:2" json:"key"`     // Logical key: "verify", "order_placed"
	Provider string `gorm:"type:varchar(50);not null;uniqueIndex:idx_sms_templates_key_provider_tenant,priority:3" json:"provider"` // Provider name: "kavenegar", "twilio"

	// Provider-specific template
	ProviderTemplate string `gorm:"type:varchar(255)" json:"providerTemplate,omitempty"` // For lookup: template name on provider panel
	MessageTemplate  string `gorm:"type:text" json:"messageTemplate,omitempty"`          // For text: message with {{placeholders}}

	// Token mapping (maps our keys to provider token names)
	// Example: {"code": "token", "name": "token2", "amount": "token3"}
	TokenMappingRaw string `gorm:"column:token_mapping;type:jsonb;default:'{}'" json:"-"`

	// Metadata
	Description string `gorm:"type:text" json:"description,omitempty"`
	IsActive    bool   `gorm:"not null;default:true" json:"isActive"`

	// Audit fields
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// TokenMapping returns the parsed token mapping
func (st *SMSTemplate) TokenMapping() map[string]string {
	mapping := make(map[string]string)
	if st.TokenMappingRaw != "" {
		_ = json.Unmarshal([]byte(st.TokenMappingRaw), &mapping)
	}
	return mapping
}

// SetTokenMapping sets the token mapping from a map
func (st *SMSTemplate) SetTokenMapping(mapping map[string]string) error {
	data, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	st.TokenMappingRaw = string(data)
	return nil
}

// MapTokens maps input data keys to provider token names using the token mapping
// For example: {"code": "123456"} with mapping {"code": "token"} returns {"token": "123456"}
func (st *SMSTemplate) MapTokens(data map[string]string) map[string]string {
	mapping := st.TokenMapping()
	result := make(map[string]string)

	for key, value := range data {
		if providerKey, ok := mapping[key]; ok {
			// Use mapped key
			result[providerKey] = value
		} else {
			// Pass through as-is if no mapping
			result[key] = value
		}
	}

	return result
}

// BeforeCreate hook to generate UUID if not set
func (st *SMSTemplate) BeforeCreate(tx *gorm.DB) error {
	if st.ID == uuid.Nil {
		st.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (SMSTemplate) TableName() string {
	return "sms_templates"
}

// SMSSendType indicates how the SMS should be sent
type SMSSendType string

const (
	SMSSendTypeLookup SMSSendType = "lookup" // Template-based (Kavenegar lookup)
	SMSSendTypeText   SMSSendType = "text"   // Plain text SMS
)
