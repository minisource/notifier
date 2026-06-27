package provider

import (
	"context"
	"fmt"
	"log"
	"time"
)

// MockProvider is a mock implementation for all channels.
// It logs send attempts and always succeeds — useful for development and testing.
type MockProvider struct {
	channel   Channel
	name      string
	shouldFail bool // When true, simulates provider failure
}

// NewMockProvider creates a new mock provider for the given channel
func NewMockProvider(channel Channel, opts ...MockOption) *MockProvider {
	p := &MockProvider{
		channel:    channel,
		name:       fmt.Sprintf("mock_%s", channel),
		shouldFail: false,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// MockOption configures a MockProvider
type MockOption func(*MockProvider)

// WithMockName sets a custom provider name
func WithMockName(name string) MockOption {
	return func(p *MockProvider) {
		p.name = name
	}
}

// WithMockFailure makes the mock provider always fail
func WithMockFailure() MockOption {
	return func(p *MockProvider) {
		p.shouldFail = true
	}
}

// Channel returns the channel this provider handles
func (p *MockProvider) Channel() Channel {
	return p.channel
}

// Name returns the provider name
func (p *MockProvider) Name() string {
	return p.name
}

// Send delivers a mock message
func (p *MockProvider) Send(ctx context.Context, msg *Message) (*SendResult, error) {
	if p.shouldFail {
		log.Printf("[MOCK PROVIDER] %s/%s — SIMULATED FAILURE for %s", p.name, p.channel, msg.To)
		return NewFailureResult(p.name, p.channel, ErrorProviderError, "simulated provider failure"), nil
	}

	// Log the mock send
	recipientHint := msg.To
	if len(recipientHint) > 20 {
		recipientHint = recipientHint[:20] + "..."
	}
	log.Printf("[MOCK PROVIDER] %s/%s — Sending to %s: subject=%q body=%q metadata=%v",
		p.name, p.channel, recipientHint, truncate(msg.Subject, 50), truncate(msg.Body, 100), msg.Metadata)

	// Generate a mock provider message ID
	mockID := fmt.Sprintf("mock_%s_%d", p.channel, time.Now().UnixNano())

	return NewSuccessResult(p.name, p.channel, mockID), nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
