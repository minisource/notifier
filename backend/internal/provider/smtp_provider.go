package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/minisource/notifier/internal/platform/email"
)

// SMTPProvider sends emails via SMTP using the existing email.SMTPClient.
// It wraps the existing implementation into the unified Provider interface.
type SMTPProvider struct {
	config *email.ProviderConfig
	client *email.SMTPClient
}

// NewSMTPProvider creates a new SMTP email provider from configuration
func NewSMTPProvider(config *email.ProviderConfig) (*SMTPProvider, error) {
	if config.Host == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}
	client, err := email.NewSMTPClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP client: %w", err)
	}
	return &SMTPProvider{
		config: config,
		client: client,
	}, nil
}

// NewSMTPProviderFromEnv creates a new SMTP provider from direct parameters
func NewSMTPProviderFromEnv(host string, port int, username, password, from, fromName string, useTLS bool) (*SMTPProvider, error) {
	cfg := &email.ProviderConfig{
		Provider: "smtp",
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		FromName: fromName,
		UseTLS:   useTLS,
	}
	return NewSMTPProvider(cfg)
}

// Channel returns the email channel
func (p *SMTPProvider) Channel() Channel {
	return ChannelEmail
}

// Name returns the provider name
func (p *SMTPProvider) Name() string {
	return "smtp"
}

// Send delivers an email via SMTP
func (p *SMTPProvider) Send(ctx context.Context, msg *Message) (*SendResult, error) {
	if msg.To == "" {
		return NewFailureResult(p.Name(), p.Channel(), ErrorInvalidRecipient, "recipient email is required"), nil
	}

	isHTML := msg.IsHTML
	if !isHTML && len(msg.Body) > 0 && (msg.Body[0] == '<' || (len(msg.Body) > 5 && msg.Body[:5] == "<!DOC")) {
		isHTML = true
	}

	if err := p.client.SendEmail(msg.To, msg.Subject, msg.Body, isHTML); err != nil {
		return NewFailureResult(p.Name(), p.Channel(), classifyEmailError(err), err.Error()), nil
	}

	providerID := fmt.Sprintf("smtp_%s", msg.NotificationID)
	return NewSuccessResult(p.Name(), p.Channel(), providerID), nil
}

func classifyEmailError(err error) ProviderErrorCode {
	errStr := err.Error()
	switch {
	case containsAny(errStr, "timed out", "timeout", "connection refused", "no such host"):
		return ErrorTimeout
	case containsAny(errStr, "rate limit", "too many", "throttle"):
		return ErrorRateLimited
	case containsAny(errStr, "authentication", "auth", "username and password not accepted"):
		return ErrorInvalidConfig
	case containsAny(errStr, "address", "recipient", "invalid"):
		return ErrorInvalidRecipient
	default:
		return ErrorProviderError
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
