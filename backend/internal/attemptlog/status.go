package attemptlog

import (
	"strings"

	"github.com/minisource/notifier/internal/provider"
)

// Status is the provider-neutral attempt lifecycle status.
type Status string

const (
	StatusQueued      Status = "queued"
	StatusPreparing   Status = "preparing"
	StatusSending     Status = "sending"
	StatusAccepted    Status = "accepted" // provider accepted the request (NOT final delivery)
	StatusPending     Status = "pending"  // accepted, awaiting async delivery status
	StatusDelivered   Status = "delivered"
	StatusFailed      Status = "failed"
	StatusRejected    Status = "rejected"   // provider rejected the request (non-retryable)
	StatusTimedOut    Status = "timed_out"
	StatusCancelled   Status = "cancelled"
	StatusBounced     Status = "bounced"
	StatusComplained  Status = "complained"
	StatusUnknown     Status = "unknown"
)

// AllStatuses returns every valid lifecycle status (for validation/UI filters).
func AllStatuses() []Status {
	return []Status{
		StatusQueued, StatusPreparing, StatusSending, StatusAccepted, StatusPending,
		StatusDelivered, StatusFailed, StatusRejected, StatusTimedOut, StatusCancelled,
		StatusBounced, StatusComplained, StatusUnknown,
	}
}

// IsValidStatus reports whether s is a known lifecycle status.
func IsValidStatus(s string) bool {
	for _, st := range AllStatuses() {
		if string(st) == s {
			return true
		}
	}
	return false
}

// NormalizedErrorKind is a coarse error taxonomy used for safe filtering and
// metrics labels (never the raw message — those go to NormalizedErrorMessage).
type NormalizedErrorKind string

const (
	ErrKindConfig        NormalizedErrorKind = "configuration"
	ErrKindRecipient     NormalizedErrorKind = "invalid_recipient"
	ErrKindContent       NormalizedErrorKind = "invalid_message"
	ErrKindProvider      NormalizedErrorKind = "provider"
	ErrKindRateLimit     NormalizedErrorKind = "rate_limited"
	ErrKindTimeout       NormalizedErrorKind = "timeout"
	ErrKindNetwork       NormalizedErrorKind = "network"
	ErrKindAuth          NormalizedErrorKind = "authentication"
	ErrKindCancelled     NormalizedErrorKind = "cancelled"
	ErrKindUnknown       NormalizedErrorKind = "unknown"
)

// NormalizeFromProviderError maps the unified provider error taxonomy
// (provider.ProviderErrorCode) onto the attempt log error kind.
func NormalizeFromProviderError(code provider.ProviderErrorCode) (NormalizedErrorKind, string) {
	switch code {
	case provider.ErrorNotConfigured, provider.ErrorInvalidConfig:
		return ErrKindConfig, string(code)
	case provider.ErrorInvalidRecipient:
		return ErrKindRecipient, string(code)
	case provider.ErrorRateLimited:
		return ErrKindRateLimit, string(code)
	case provider.ErrorTimeout:
		return ErrKindTimeout, string(code)
	case provider.ErrorServiceUnavailable, provider.ErrorNetworkError:
		return ErrKindNetwork, string(code)
	case provider.ErrorInvalidMessage, provider.ErrorTemplateNotFound:
		return ErrKindContent, string(code)
	case provider.ErrorProviderError:
		return ErrKindProvider, string(code)
	default:
		return ErrKindUnknown, string(code)
	}
}

// ClassifyErrorText heuristically classifies an arbitrary provider error string
// into a normalized kind + code. Used for network/transport errors that arrive
// outside the unified taxonomy (e.g. dial errors, malformed responses).
func ClassifyErrorText(errText string) (NormalizedErrorKind, string) {
	lower := strings.ToLower(errText)
	switch {
	case containsAny(lower, "timeout", "timed out", "deadline exceeded"):
		return ErrKindTimeout, "timeout"
	case containsAny(lower, "connection refused", "connection reset", "no such host", "dial tcp", "network is unreachable", "broken pipe"):
		return ErrKindNetwork, "network_error"
	case containsAny(lower, "rate limit", "too many requests", "throttl"):
		return ErrKindRateLimit, "rate_limited"
	case containsAny(lower, "unauthorized", "forbidden", "invalid api key", "authentication failed", "access denied", "401"):
		return ErrKindAuth, "authentication"
	case containsAny(lower, "malformed", "invalid response", "unmarshal", "unexpected end of json"):
		return ErrKindUnknown, "malformed_response"
	case containsAny(lower, "not configured", "no provider", "unsupported"):
		return ErrKindConfig, "not_configured"
	default:
		return ErrKindUnknown, "unknown"
	}
}

func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}
