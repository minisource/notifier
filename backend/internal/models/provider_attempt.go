package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProviderAttempt represents one durable record of a single outbound provider
// request during the delivery lifecycle of a notification. It is the
// authoritative, searchable history for provider attempts (retries and
// fallbacks are distinct rows linked via attempt_number/fallback_sequence).
//
// Security: this table NEVER stores raw secrets, full message bodies, or full
// recipient values. All request/response payloads are sanitized (redacted,
// masked, truncated) before persistence.
type ProviderAttempt struct {
	ID     uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID *uuid.UUID `gorm:"type:uuid;index" json:"tenantId,omitempty"`

	// References
	NotificationID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"notificationId"`
	ProviderAccountID *uuid.UUID `gorm:"type:uuid;index" json:"providerAccountId,omitempty"` // providers.id when known
	ParentAttemptID *uuid.UUID `gorm:"type:uuid;index" json:"parentAttemptId,omitempty"`     // linked retry/fallback chain

	// Context
	Channel          string `gorm:"type:varchar(20);not null;index" json:"channel"` // sms | email | push | in_app
	Provider         string `gorm:"type:varchar(100);not null;index" json:"provider"`
	AttemptNumber    int    `gorm:"not null;default:1" json:"attemptNumber"`
	FallbackSequence int    `gorm:"not null;default:0" json:"fallbackSequence"` // 0 = primary, 1+ = fallback provider order

	// Status model (provider-neutral, see attemptlog/status.go)
	Status                  string `gorm:"type:varchar(30);not null;index" json:"status"` // queued|preparing|sending|accepted|pending|delivered|failed|rejected|timed_out|cancelled|unknown
	ProviderStatus          string `gorm:"type:varchar(100)" json:"providerStatus,omitempty"`
	ProviderMessageID       string `gorm:"type:varchar(255);index" json:"providerMessageId,omitempty"`
	ProviderErrorCode       string `gorm:"type:varchar(100)" json:"providerErrorCode,omitempty"`
	NormalizedErrorKind     string `gorm:"type:varchar(50)" json:"normalizedErrorKind,omitempty"`
	NormalizedErrorCode     string `gorm:"type:varchar(100)" json:"normalizedErrorCode,omitempty"`
	NormalizedErrorMessage  string `gorm:"type:text" json:"normalizedErrorMessage,omitempty"`
	Retryable               bool   `gorm:"not null;default:false" json:"retryable"`

	// Sanitized outbound request
	RequestMethod          string `gorm:"type:varchar(10)" json:"requestMethod,omitempty"`
	RequestURLSanitized    string `gorm:"type:text" json:"requestUrlSanitized,omitempty"`
	RequestHeadersSanitized string `gorm:"type:jsonb" json:"requestHeadersSanitized,omitempty"`
	RequestBodySanitized   string `gorm:"type:text" json:"requestBodySanitized,omitempty"`
	RequestSizeBytes       int    `json:"requestSizeBytes,omitempty"`

	// Sanitized provider response
	ResponseStatusCode     int    `gorm:"index" json:"responseStatusCode,omitempty"`
	ResponseHeadersSanitized string `gorm:"type:jsonb" json:"responseHeadersSanitized,omitempty"`
	ResponseBodySanitized  string `gorm:"type:text" json:"responseBodySanitized,omitempty"`
	ResponseSizeBytes      int    `json:"responseSizeBytes,omitempty"`

	// Body capture control (truncation metadata)
	BodyTruncated       bool `gorm:"not null;default:false" json:"bodyTruncated"`
	OriginalSizeBytes   int  `json:"originalSizeBytes,omitempty"`
	CapturedSizeBytes   int  `json:"capturedSizeBytes,omitempty"`

	// Content policy: hash + bounded preview only (never the full message)
	ContentHash  string `gorm:"type:varchar(128)" json:"contentHash,omitempty"`
	BodyPreview  string `gorm:"type:text" json:"bodyPreview,omitempty"`

	// Masked recipient (never the full phone/email — see attemptlog masking)
	RecipientMasked string `gorm:"type:varchar(255)" json:"recipientMasked,omitempty"`

	// Timing
	QueuedAt     time.Time  `gorm:"index" json:"queuedAt"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
	DurationMs   int64      `gorm:"index" json:"durationMs,omitempty"`
	TimeoutMs    int        `json:"timeoutMs,omitempty"`

	// Diagnostics correlation
	RequestID     string `gorm:"type:varchar(64);index" json:"requestId,omitempty"`
	CorrelationID string `gorm:"type:varchar(64);index" json:"correlationId,omitempty"`
	TraceID       string `gorm:"type:varchar(64)" json:"traceId,omitempty"`
	SpanID        string `gorm:"type:varchar(64)" json:"spanId,omitempty"`

	// Audit
	CreatedAt time.Time      `gorm:"not null;default:now();index" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// BeforeCreate hook to generate UUID if not set and default JSON values.
func (a *ProviderAttempt) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.RequestHeadersSanitized == "" {
		a.RequestHeadersSanitized = "{}"
	}
	if a.ResponseHeadersSanitized == "" {
		a.ResponseHeadersSanitized = "{}"
	}
	return nil
}

// TableName specifies the table name.
func (ProviderAttempt) TableName() string { return "notification_provider_attempts" }

// ProviderAttemptEvent represents a single lifecycle event belonging to a
// provider attempt (append-only timeline).
type ProviderAttemptEvent struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AttemptID uuid.UUID `gorm:"type:uuid;not null;index" json:"attemptId"`

	EventType        string `gorm:"type:varchar(50);not null" json:"eventType"` // attempt_created|request_started|response_received|request_failed|request_timed_out|retry_scheduled|fallback_scheduled|accepted|delivered|delivery_failed|...
	PreviousStatus   string `gorm:"type:varchar(30)" json:"previousStatus,omitempty"`
	NewStatus        string `gorm:"type:varchar(30)" json:"newStatus,omitempty"`
	EventPayloadSanitized string `gorm:"type:jsonb" json:"eventPayloadSanitized,omitempty"`
	Source           string `gorm:"type:varchar(50)" json:"source,omitempty"` // worker | adapter | webhook | system

	RequestID     string `gorm:"type:varchar(64);index" json:"requestId,omitempty"`
	CorrelationID string `gorm:"type:varchar(64);index" json:"correlationId,omitempty"`
	TraceID       string `gorm:"type:varchar(64)" json:"traceId,omitempty"`

	OccurredAt time.Time `gorm:"not null;default:now();index" json:"occurredAt"`
}

// BeforeCreate hook to generate UUID if not set.
func (e *ProviderAttemptEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.EventPayloadSanitized == "" {
		e.EventPayloadSanitized = "{}"
	}
	return nil
}

// TableName specifies the table name.
func (ProviderAttemptEvent) TableName() string { return "notification_provider_attempt_events" }
