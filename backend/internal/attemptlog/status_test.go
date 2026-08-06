package attemptlog

import (
	"testing"

	"github.com/minisource/notifier/internal/provider"
)

func TestAllStatusesValid(t *testing.T) {
	for _, s := range AllStatuses() {
		if !IsValidStatus(string(s)) {
			t.Errorf("status %q not recognized as valid", s)
		}
	}
	if IsValidStatus("not-a-status") {
		t.Errorf("invalid status should be rejected")
	}
}

func TestNormalizeFromProviderError(t *testing.T) {
	cases := []struct {
		code provider.ProviderErrorCode
		want NormalizedErrorKind
	}{
		{provider.ErrorNotConfigured, ErrKindConfig},
		{provider.ErrorInvalidConfig, ErrKindConfig},
		{provider.ErrorInvalidRecipient, ErrKindRecipient},
		{provider.ErrorRateLimited, ErrKindRateLimit},
		{provider.ErrorTimeout, ErrKindTimeout},
		{provider.ErrorServiceUnavailable, ErrKindNetwork},
		{provider.ErrorNetworkError, ErrKindNetwork},
		{provider.ErrorInvalidMessage, ErrKindContent},
		{provider.ErrorTemplateNotFound, ErrKindContent},
		{provider.ErrorProviderError, ErrKindProvider},
	}
	for _, c := range cases {
		kind, code := NormalizeFromProviderError(c.code)
		if kind != c.want {
			t.Errorf("code %q: want kind %q, got %q", c.code, c.want, kind)
		}
		if code != string(c.code) {
			t.Errorf("code %q: normalized code mismatch %q", c.code, code)
		}
	}
}

func TestClassifyErrorText(t *testing.T) {
	cases := []struct {
		text string
		want NormalizedErrorKind
	}{
		{"dial tcp 10.0.0.1:443: connect: connection refused", ErrKindNetwork},
		{"request timed out after 10s", ErrKindTimeout},
		{"rate limit exceeded: too many requests", ErrKindRateLimit},
		{"unauthorized: invalid API key", ErrKindAuth},
		{"malformed response: unexpected end of JSON input", ErrKindUnknown},
		{"provider not configured", ErrKindConfig},
		{"some unknown error happened", ErrKindUnknown},
	}
	for _, c := range cases {
		kind, _ := ClassifyErrorText(c.text)
		if kind != c.want {
			t.Errorf("text %q: want %q, got %q", c.text, c.want, kind)
		}
	}
}
