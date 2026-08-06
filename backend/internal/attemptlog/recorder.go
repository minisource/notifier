package attemptlog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
)

// AttemptRepository is the persistence contract the recorder needs. It is
// satisfied by repository.ProviderAttemptRepository.
type AttemptRepository interface {
	CreateAttempt(ctx context.Context, attempt *models.ProviderAttempt) error
	UpdateAttempt(ctx context.Context, attempt *models.ProviderAttempt) error
	AddEvent(ctx context.Context, event *models.ProviderAttemptEvent) error
}

// StartInput carries everything needed to open a new provider attempt.
type StartInput struct {
	NotificationID  uuid.UUID
	TenantID        *uuid.UUID
	ProviderAccountID *uuid.UUID
	ParentAttemptID *uuid.UUID
	Channel         string
	Provider        string
	AttemptNumber   int
	FallbackSequence int
	TimeoutMs       int

	// Sanitized request metadata (caller must pass already-sanitized values).
	RequestMethod           string
	RequestURLSanitized     string
	RequestHeadersSanitized map[string]string
	RequestBodySanitized    string
	RequestSizeBytes        int

	// Content policy: hash + preview (never full content).
	ContentHash  string
	BodyPreview  string

	// Masked recipient (masking must happen in the caller, before this struct).
	RecipientMasked string

	RequestID     string
	CorrelationID string
	TraceID       string
	SpanID        string
}

// FinishInput carries the outcome of provider execution.
type FinishInput struct {
	Status                  Status
	ProviderStatus          string
	ProviderMessageID       string
	ProviderErrorCode       string
	NormalizedErrorKind     string
	NormalizedErrorCode     string
	NormalizedErrorMessage  string
	Retryable               bool
	ResponseStatusCode      int
	ResponseHeadersSanitized map[string]string
	ResponseBodySanitized   string
	ResponseSizeBytes       int
	BodyTruncated           bool
	OriginalSizeBytes       int
	CapturedSizeBytes       int
	StartedAt               time.Time
}

// Recorder is the single shared provider-execution boundary used by every
// channel adapter (SMS/Email/Push). It persists a durable attempt record plus
// lifecycle events with strict redaction. Persistence failures never block
// delivery: they are logged as observable errors only.
type Recorder struct {
	repo   AttemptRepository
	logger logging.Logger
	opts   RedactionOptions
	// Enabled toggles the whole feature from config (ProviderLogs.Enabled).
	Enabled bool
}

// NewRecorder creates a Recorder. Pass nil logger to disable logging.
func NewRecorder(repo AttemptRepository, logger logging.Logger, opts RedactionOptions, enabled bool) *Recorder {
	if opts.MaxBodyBytes == 0 {
		opts = DefaultRedactionOptions()
	}
	return &Recorder{repo: repo, logger: logger, opts: opts, Enabled: enabled}
}

// NewRequestID returns a random hex request id.
func NewRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return uuid.NewString()
	}
	return hex.EncodeToString(b)
}

// Start persists a new attempt in "queued"/"preparing" state and records the
// attempt_created event. Returns the attempt (with ID) or nil when the feature
// is disabled.
func (r *Recorder) Start(ctx context.Context, in StartInput) *models.ProviderAttempt {
	if r == nil || !r.Enabled || r.repo == nil {
		return nil
	}
	now := time.Now().UTC()
	attempt := &models.ProviderAttempt{
		NotificationID:        in.NotificationID,
		TenantID:              in.TenantID,
		ProviderAccountID:     in.ProviderAccountID,
		ParentAttemptID:       in.ParentAttemptID,
		Channel:               in.Channel,
		Provider:              in.Provider,
		AttemptNumber:         in.AttemptNumber,
		FallbackSequence:      in.FallbackSequence,
		Status:                string(StatusQueued),
		RequestMethod:         in.RequestMethod,
		RequestURLSanitized:   in.RequestURLSanitized,
		RequestHeadersSanitized: mustJSON(in.RequestHeadersSanitized),
		RequestBodySanitized:  in.RequestBodySanitized,
		RequestSizeBytes:      in.RequestSizeBytes,
		ContentHash:           in.ContentHash,
		BodyPreview:           in.BodyPreview,
		RecipientMasked:       in.RecipientMasked,
		QueuedAt:              now,
		TimeoutMs:             in.TimeoutMs,
		RequestID:             in.RequestID,
		CorrelationID:         in.CorrelationID,
		TraceID:               in.TraceID,
		SpanID:                in.SpanID,
	}
	if attempt.RequestID == "" {
		attempt.RequestID = NewRequestID()
	}
	if attempt.CorrelationID == "" {
		attempt.CorrelationID = in.NotificationID.String()
	}

	if err := r.repo.CreateAttempt(ctx, attempt); err != nil {
		r.logErr("failed to persist provider attempt", in.NotificationID, in.Provider, err)
		return nil
	}
	r.recordEvent(ctx, attempt, "attempt_created", "", string(StatusQueued), map[string]interface{}{"channel": in.Channel, "provider": in.Provider, "attemptNumber": in.AttemptNumber, "fallbackSequence": in.FallbackSequence})
	return attempt
}

// MarkSending transitions the attempt to "sending" and records request_started.
func (r *Recorder) MarkSending(ctx context.Context, attempt *models.ProviderAttempt) {
	if attempt == nil {
		return
	}
	now := time.Now().UTC()
	attempt.Status = string(StatusSending)
	attempt.StartedAt = &now
	if err := r.repo.UpdateAttempt(ctx, attempt); err != nil {
		r.logErr("failed to mark attempt sending", attempt.NotificationID, attempt.Provider, err)
		return
	}
	r.recordEvent(ctx, attempt, "request_started", "", string(StatusSending), map[string]interface{}{"requestId": attempt.RequestID})
}

// UpdateRequest replaces the sanitized request metadata once the adapter knows
// the REAL outbound request (method/URL/body) — e.g. after the provider client
// is constructed and the exact endpoint + form body are determined. The values
// must already be sanitized/masked by the adapter (see RequestDescriber).
func (r *Recorder) UpdateRequest(ctx context.Context, attempt *models.ProviderAttempt, method, urlSanitized, bodySanitized string, sizeBytes int) {
	if attempt == nil {
		return
	}
	attempt.RequestMethod = method
	attempt.RequestURLSanitized = urlSanitized
	attempt.RequestBodySanitized = bodySanitized
	attempt.RequestSizeBytes = sizeBytes
	if err := r.repo.UpdateAttempt(ctx, attempt); err != nil {
		r.logErr("failed to update provider attempt request", attempt.NotificationID, attempt.Provider, err)
	}
}

// Finish finalizes the attempt with the outcome and records the terminal event.
func (r *Recorder) Finish(ctx context.Context, attempt *models.ProviderAttempt, in FinishInput) {
	if attempt == nil {
		return
	}
	now := time.Now().UTC()
	attempt.Status = string(in.Status)
	attempt.ProviderStatus = in.ProviderStatus
	attempt.ProviderMessageID = in.ProviderMessageID
	attempt.ProviderErrorCode = in.ProviderErrorCode
	attempt.NormalizedErrorKind = in.NormalizedErrorKind
	attempt.NormalizedErrorCode = in.NormalizedErrorCode
	attempt.NormalizedErrorMessage = truncateErr(in.NormalizedErrorMessage)
	attempt.Retryable = in.Retryable
	attempt.ResponseStatusCode = in.ResponseStatusCode
	attempt.ResponseHeadersSanitized = mustJSON(in.ResponseHeadersSanitized)
	attempt.ResponseBodySanitized = in.ResponseBodySanitized
	attempt.ResponseSizeBytes = in.ResponseSizeBytes
	attempt.BodyTruncated = in.BodyTruncated
	attempt.OriginalSizeBytes = in.OriginalSizeBytes
	attempt.CapturedSizeBytes = in.CapturedSizeBytes
	attempt.CompletedAt = &now
	if !in.StartedAt.IsZero() {
		attempt.DurationMs = now.Sub(in.StartedAt).Milliseconds()
	}

	if err := r.repo.UpdateAttempt(ctx, attempt); err != nil {
		r.logErr("failed to finalize provider attempt", attempt.NotificationID, attempt.Provider, err)
	}

	eventType := eventTypeForStatus(in.Status)
	payload := map[string]interface{}{}
	if in.ProviderMessageID != "" {
		payload["providerMessageId"] = in.ProviderMessageID
	}
	if in.ProviderStatus != "" {
		payload["providerStatus"] = in.ProviderStatus
	}
	if in.ResponseStatusCode != 0 {
		payload["responseStatusCode"] = in.ResponseStatusCode
	}
	if in.NormalizedErrorCode != "" {
		payload["errorCode"] = in.NormalizedErrorCode
	}
	if in.Retryable {
		payload["retryable"] = true
	}
	r.recordEvent(ctx, attempt, eventType, string(StatusSending), string(in.Status), payload)
}

// RecordRetryScheduled records that a retry was scheduled for this attempt.
func (r *Recorder) RecordRetryScheduled(ctx context.Context, attempt *models.ProviderAttempt, nextRetryAt time.Time, errText string) {
	if attempt == nil {
		return
	}
	attempt.Status = string(StatusFailed)
	r.recordEvent(ctx, attempt, "retry_scheduled", "", string(StatusFailed), map[string]interface{}{
		"nextRetryAt": nextRetryAt.Format(time.RFC3339),
		"error":       truncateErr(errText),
	})
}

// RecordFallbackScheduled records that a fallback provider was attempted next.
func (r *Recorder) RecordFallbackScheduled(ctx context.Context, attempt *models.ProviderAttempt, fallbackProvider string) {
	if attempt == nil {
		return
	}
	r.recordEvent(ctx, attempt, "fallback_scheduled", "", string(StatusFailed), map[string]interface{}{
		"fallbackProvider": fallbackProvider,
	})
}

// recordEvent appends an event row. Failures are logged, never returned.
func (r *Recorder) recordEvent(ctx context.Context, attempt *models.ProviderAttempt, eventType, prev, next string, payload map[string]interface{}) {
	ev := &models.ProviderAttemptEvent{
		AttemptID:            attempt.ID,
		EventType:            eventType,
		PreviousStatus:       prev,
		NewStatus:            next,
		EventPayloadSanitized: mustJSON(payload),
		Source:               "adapter",
		RequestID:            attempt.RequestID,
		CorrelationID:        attempt.CorrelationID,
		TraceID:              attempt.TraceID,
		OccurredAt:           time.Now().UTC(),
	}
	if err := r.repo.AddEvent(ctx, ev); err != nil {
		r.logErr("failed to persist attempt event", attempt.NotificationID, attempt.Provider, err)
	}
}

func (r *Recorder) logErr(msg string, notificationID uuid.UUID, providerName string, err error) {
	if r.logger == nil {
		return
	}
	r.logger.Error(logging.General, logging.Insert, msg, map[logging.ExtraKey]interface{}{
		"notificationId": notificationID,
		"provider":       providerName,
		"error":          err.Error(),
	})
}

func eventTypeForStatus(s Status) string {
	switch s {
	case StatusAccepted:
		return "provider_accepted"
	case StatusDelivered:
		return "delivered"
	case StatusRejected:
		return "provider_rejected"
	case StatusTimedOut:
		return "request_timed_out"
	case StatusCancelled:
		return "request_cancelled"
	case StatusFailed:
		return "request_failed"
	case StatusPending:
		return "delivery_pending"
	case StatusBounced:
		return "bounced"
	case StatusComplained:
		return "complained"
	case StatusUnknown:
		return "malformed_response"
	default:
		return "response_received"
	}
}

func mustJSON(v interface{}) string {
	if v == nil {
		return "{}"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func truncateErr(msg string) string {
	if len(msg) > 1000 {
		return msg[:1000] + "..."
	}
	return msg
}

// SanitizeRequestBody is a helper for adapters: redacts + bounds a raw request
// body. Returns (sanitized, truncated, original, captured).
func SanitizeRequestBody(raw string, opts RedactionOptions) (string, bool, int, int) {
	return SanitizeBody(raw, opts)
}

// SanitizeResponseBody is a helper for adapters: redacts + bounds a raw
// response body. Returns (sanitized, truncated, original, captured).
func SanitizeResponseBody(raw string, opts RedactionOptions) (string, bool, int, int) {
	return SanitizeBody(raw, opts)
}
