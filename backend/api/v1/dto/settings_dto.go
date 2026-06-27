package dto

// NotificationSettingsResponse represents the full notification settings configuration
type NotificationSettingsResponse struct {
	// Default providers per channel
	DefaultEmailProviderID  *string `json:"defaultEmailProviderId,omitempty"`
	DefaultSMSProviderID    *string `json:"defaultSmsProviderId,omitempty"`
	DefaultPushProviderID   *string `json:"defaultPushProviderId,omitempty"`
	DefaultWebhookProviderID *string `json:"defaultWebhookProviderId,omitempty"`

	// Enabled channels
	EnabledChannels EnabledChannels `json:"enabledChannels"`

	// Retry policy
	RetryPolicy RetryPolicy `json:"retryPolicy"`

	// Rate limits
	RateLimit RateLimit `json:"rateLimit"`

	// Quiet hours
	QuietHours *QuietHoursConfig `json:"quietHours,omitempty"`

	// Retention
	RetentionDays int `json:"retentionDays"`
}

// EnabledChannels represents which notification channels are enabled
type EnabledChannels struct {
	Email   bool `json:"email"`
	SMS     bool `json:"sms"`
	Push    bool `json:"push"`
	Webhook bool `json:"webhook"`
	InApp   bool `json:"inApp"`
}

// RetryPolicy represents the notification retry configuration
type RetryPolicy struct {
	Enabled              bool   `json:"enabled"`
	MaxAttempts          int    `json:"maxAttempts"`
	BackoffStrategy      string `json:"backoffStrategy"` // fixed, linear, exponential
	InitialDelaySeconds  int    `json:"initialDelaySeconds"`
	MaxDelaySeconds      int    `json:"maxDelaySeconds"`
}

// RateLimit represents the notification rate limiting configuration
type RateLimit struct {
	Enabled   bool `json:"enabled"`
	PerMinute int  `json:"perMinute"`
	PerHour   int  `json:"perHour"`
}

// QuietHoursConfig represents quiet hours configuration
type QuietHoursConfig struct {
	Enabled  bool   `json:"enabled"`
	Timezone string `json:"timezone"`
	Start    string `json:"start"` // HH:mm
	End      string `json:"end"`   // HH:mm
}

// UpdateNotificationSettingsRequest represents a request to update notification settings
type UpdateNotificationSettingsRequest struct {
	DefaultEmailProviderID  *string                     `json:"defaultEmailProviderId,omitempty"`
	DefaultSMSProviderID    *string                     `json:"defaultSmsProviderId,omitempty"`
	DefaultPushProviderID   *string                     `json:"defaultPushProviderId,omitempty"`
	DefaultWebhookProviderID *string                    `json:"defaultWebhookProviderId,omitempty"`
	EnabledChannels         *EnabledChannels            `json:"enabledChannels,omitempty"`
	RetryPolicy             *RetryPolicy                `json:"retryPolicy,omitempty"`
	RateLimit               *RateLimit                  `json:"rateLimit,omitempty"`
	QuietHours              *QuietHoursConfig           `json:"quietHours,omitempty"`
	RetentionDays           *int                        `json:"retentionDays,omitempty"`
}

// DefaultNotificationSettings returns the default notification settings
func DefaultNotificationSettings() *NotificationSettingsResponse {
	return &NotificationSettingsResponse{
		EnabledChannels: EnabledChannels{
			Email:   true,
			SMS:     true,
			Push:    true,
			Webhook: true,
			InApp:   true,
		},
		RetryPolicy: RetryPolicy{
			Enabled:              true,
			MaxAttempts:          3,
			BackoffStrategy:      "exponential",
			InitialDelaySeconds:  60,
			MaxDelaySeconds:      3600,
		},
		RateLimit: RateLimit{
			Enabled:   true,
			PerMinute: 100,
			PerHour:   1000,
		},
		QuietHours:    nil,
		RetentionDays: 90,
	}
}
