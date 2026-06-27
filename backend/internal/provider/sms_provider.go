package provider

import (
	"context"
	"fmt"

	"github.com/minisource/notifier/internal/platform/sms"
	smsproviders "github.com/minisource/notifier/internal/platform/sms/platforms"
)

// SMSProvider sends SMS messages using a configured provider client.
// It wraps the existing sms.SmsClient into the unified Provider interface.
type SMSProvider struct {
	name   string
	client smsproviders.SmsClient
}

// NewSMSProviderFromConfig creates an SMS provider from a configuration string
func NewSMSProviderFromConfig(configJSON string) (*SMSProvider, error) {
	cfg, err := sms.ParseProviderConfig(configJSON)
	if err != nil {
		return nil, fmt.Errorf("invalid SMS config: %w", err)
	}

	client, err := sms.NewClientFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMS client: %w", err)
	}

	return &SMSProvider{
		name:   cfg.Provider,
		client: client,
	}, nil
}

// NewSMSProvider creates an SMS provider from a provider name and client
func NewSMSProvider(name string, client smsproviders.SmsClient) *SMSProvider {
	return &SMSProvider{
		name:   name,
		client: client,
	}
}

// Channel returns the SMS channel
func (p *SMSProvider) Channel() Channel {
	return ChannelSMS
}

// Name returns the provider name
func (p *SMSProvider) Name() string {
	return p.name
}

// Send delivers an SMS via the configured provider
func (p *SMSProvider) Send(ctx context.Context, msg *Message) (*SendResult, error) {
	if msg.To == "" {
		return NewFailureResult(p.Name(), p.Channel(), ErrorInvalidRecipient, "phone number is required"), nil
	}

	// Build params for the existing SMS client interface
	params := make(map[string]string)
	if msg.Subject != "" {
		params["template"] = msg.Subject
	}
	params["body"] = msg.Body
	params["message"] = msg.Body
	params["code"] = msg.Body
	params["token"] = msg.Body

	// Copy metadata into params
	for k, v := range msg.Metadata {
		params[k] = v
	}

	// Send to recipient
	if err := p.client.SendMessage(params, msg.To); err != nil {
		return NewFailureResult(p.Name(), p.Channel(), classifySMSError(err), err.Error()), nil
	}

	providerID := fmt.Sprintf("sms_%s_%s", p.name, msg.NotificationID)
	return NewSuccessResult(p.Name(), p.Channel(), providerID), nil
}

func classifySMSError(err error) ProviderErrorCode {
	errStr := err.Error()
	switch {
	case containsAny(errStr, "timed out", "timeout", "connection refused", "no such host"):
		return ErrorTimeout
	case containsAny(errStr, "rate limit", "too many requests", "throttle"):
		return ErrorRateLimited
	case containsAny(errStr, "credential", "api key", "api_key", "unauthorized", "invalid"):
		return ErrorInvalidConfig
	case containsAny(errStr, "not found", "no template", "template"):
		return ErrorTemplateNotFound
	case containsAny(errStr, "invalid number", "phone", "recipient"):
		return ErrorInvalidRecipient
	default:
		return ErrorProviderError
	}
}
