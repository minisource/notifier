package provider

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Channel represents a notification delivery channel
type Channel string

const (
	ChannelSMS   Channel = "sms"
	ChannelEmail Channel = "email"
	ChannelPush  Channel = "push"
	ChannelInApp Channel = "in_app"
)

// AllChannels returns all supported channels
func AllChannels() []Channel {
	return []Channel{ChannelSMS, ChannelEmail, ChannelPush, ChannelInApp}
}

// ParseChannel parses a channel string, returning an error if invalid
func ParseChannel(s string) (Channel, error) {
	switch Channel(strings.ToLower(s)) {
	case ChannelSMS, ChannelEmail, ChannelPush, ChannelInApp:
		return Channel(strings.ToLower(s)), nil
	default:
		return "", fmt.Errorf("invalid channel: %q (valid: sms, email, push, in_app)", s)
	}
}

// SendStatus represents the status of a send operation
type SendStatus string

const (
	SendStatusSuccess SendStatus = "success"
	SendStatusFailed  SendStatus = "failed"
)

// Message represents a notification message to be sent via a provider
type Message struct {
	// Recipient information
	To      string `json:"to"`      // Phone number, email address, device token, or user ID
	Subject string `json:"subject"` // Email subject or push title
	Body    string `json:"body"`    // Message body/content
	IsHTML  bool   `json:"isHtml"`  // Whether body is HTML (for email)

	// Provider-specific metadata
	Metadata map[string]string `json:"metadata,omitempty"` // e.g., template name, token values

	// Notification tracking
	NotificationID string `json:"notificationId"` // Reference to the originating notification
	UserID         string `json:"userId"`          // Target user ID
}

// SendResult represents the result of a provider send operation
type SendResult struct {
	Provider     string     `json:"provider"`
	Channel      Channel    `json:"channel"`
	ProviderID   string     `json:"providerId"`   // Provider-side message ID
	Status       SendStatus `json:"status"`        // success or failed
	Retryable    bool       `json:"retryable"`     // Whether this failure can be retried
	ErrorCode    string     `json:"errorCode,omitempty"`    // Categorized error code
	ErrorMessage string     `json:"errorMessage,omitempty"` // Sanitized error message (no PII)
	SentAt       time.Time  `json:"sentAt"`
}

// Provider is the unified interface for all notification providers
type Provider interface {
	// Channel returns the notification channel this provider handles
	Channel() Channel

	// Name returns the provider name (e.g., "kavenegar", "smtp", "mock")
	Name() string

	// Send delivers a message through this provider
	Send(ctx context.Context, msg *Message) (*SendResult, error)
}

// ProviderErrorCode represents categorized error codes for provider failures
type ProviderErrorCode string

const (
	// Configuration errors (non-retryable)
	ErrorNotConfigured     ProviderErrorCode = "not_configured"
	ErrorInvalidConfig     ProviderErrorCode = "invalid_config"
	ErrorInvalidRecipient  ProviderErrorCode = "invalid_recipient"

	// Transient errors (retryable)
	ErrorRateLimited       ProviderErrorCode = "rate_limited"
	ErrorTimeout           ProviderErrorCode = "timeout"
	ErrorServiceUnavailable ProviderErrorCode = "service_unavailable"
	ErrorNetworkError      ProviderErrorCode = "network_error"
	ErrorProviderError     ProviderErrorCode = "provider_error"

	// Content errors (non-retryable)
	ErrorInvalidMessage    ProviderErrorCode = "invalid_message"
	ErrorTemplateNotFound  ProviderErrorCode = "template_not_found"
)

// IsRetryable determines whether a provider error is retryable based on its error code
func IsRetryable(code ProviderErrorCode) bool {
	switch code {
	case ErrorRateLimited, ErrorTimeout, ErrorServiceUnavailable, ErrorNetworkError, ErrorProviderError:
		return true
	default:
		return false
	}
}

// NewSuccessResult creates a successful SendResult
func NewSuccessResult(providerName string, channel Channel, providerID string) *SendResult {
	return &SendResult{
		Provider:   providerName,
		Channel:    channel,
		ProviderID: providerID,
		Status:     SendStatusSuccess,
		Retryable:  false,
		SentAt:     time.Now(),
	}
}

// NewFailureResult creates a failed SendResult with error classification
func NewFailureResult(providerName string, channel Channel, errCode ProviderErrorCode, errMessage string) *SendResult {
	return &SendResult{
		Provider:     providerName,
		Channel:      channel,
		ProviderID:   "",
		Status:       SendStatusFailed,
		Retryable:    IsRetryable(errCode),
		ErrorCode:    string(errCode),
		ErrorMessage: sanitizeErrorMessage(errMessage),
		SentAt:       time.Now(),
	}
}

// sanitizeErrorMessage removes potential PII from error messages
func sanitizeErrorMessage(msg string) string {
	if msg == "" {
		return ""
	}
	// Truncate long messages
	if len(msg) > 500 {
		msg = msg[:500] + "..."
	}
	return msg
}

// ProviderNotFoundError is returned when no provider is configured for a channel
type ProviderNotFoundError struct {
	Channel Channel
}

func (e *ProviderNotFoundError) Error() string {
	return fmt.Sprintf("no provider configured for channel: %s", e.Channel)
}

// MessageValidationError is returned when a message is invalid
type MessageValidationError struct {
	Field   string
	Message string
}

func (e *MessageValidationError) Error() string {
	return fmt.Sprintf("message validation failed: %s - %s", e.Field, e.Message)
}
