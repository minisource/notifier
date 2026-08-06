package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/notifier/internal/models"
)

// == Provider Request Lifecycle Logging DTOs ==

// ProviderAttemptSummary is the compact attempt record used in list views.
// It NEVER contains request/response bodies — details are loaded on demand.
type ProviderAttemptSummary struct {
	ID               uuid.UUID  `json:"id"`
	NotificationID   uuid.UUID  `json:"notificationId"`
	ProviderAccountID *uuid.UUID `json:"providerAccountId,omitempty"`
	TenantID         *uuid.UUID `json:"tenantId,omitempty"`
	Channel          string     `json:"channel"`
	Provider         string     `json:"provider"`
	AttemptNumber    int        `json:"attemptNumber"`
	FallbackSequence int        `json:"fallbackSequence"`
	Status           string     `json:"status"`
	ProviderStatus   string     `json:"providerStatus,omitempty"`
	ProviderMessageID string    `json:"providerMessageId,omitempty"`
	RecipientMasked  string     `json:"recipientMasked,omitempty"`
	ResponseStatusCode int      `json:"responseStatusCode,omitempty"`
	DurationMs       int64      `json:"durationMs,omitempty"`
	Retryable        bool       `json:"retryable"`
	NormalizedErrorKind string `json:"normalizedErrorKind,omitempty"`
	NormalizedErrorCode string `json:"normalizedErrorCode,omitempty"`
	RequestID        string     `json:"requestId,omitempty"`
	CorrelationID    string     `json:"correlationId,omitempty"`
	TraceID          string     `json:"traceId,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
}

// ProviderAttemptDetails is the full attempt record including sanitized
// request/response payloads and lifecycle timeline.
type ProviderAttemptDetails struct {
	ProviderAttemptSummary
	ParentAttemptID *uuid.UUID `json:"parentAttemptId,omitempty"`

	// Sanitized outbound request
	RequestMethod            string            `json:"requestMethod,omitempty"`
	RequestURLSanitized      string            `json:"requestUrlSanitized,omitempty"`
	RequestHeadersSanitized  map[string]string `json:"requestHeadersSanitized,omitempty"`
	RequestBodySanitized     string            `json:"requestBodySanitized,omitempty"`
	RequestSizeBytes         int               `json:"requestSizeBytes,omitempty"`

	// Sanitized provider response
	ResponseHeadersSanitized map[string]string `json:"responseHeadersSanitized,omitempty"`
	ResponseBodySanitized    string            `json:"responseBodySanitized,omitempty"`
	ResponseSizeBytes        int               `json:"responseSizeBytes,omitempty"`

	// Body capture / content policy metadata
	BodyTruncated     bool   `json:"bodyTruncated"`
	OriginalSizeBytes int    `json:"originalSizeBytes,omitempty"`
	CapturedSizeBytes int    `json:"capturedSizeBytes,omitempty"`
	ContentHash       string `json:"contentHash,omitempty"`
	BodyPreview       string `json:"bodyPreview,omitempty"`

	ProviderErrorCode      string `json:"providerErrorCode,omitempty"`
	NormalizedErrorMessage string `json:"normalizedErrorMessage,omitempty"`

	QueuedAt   time.Time  `json:"queuedAt"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	TimeoutMs  int        `json:"timeoutMs,omitempty"`
	SpanID     string     `json:"spanId,omitempty"`

	Events []*ProviderAttemptEventResponse `json:"events,omitempty"`
}

// ProviderAttemptEventResponse is a single lifecycle timeline entry.
type ProviderAttemptEventResponse struct {
	ID                     string         `json:"id"`
	AttemptID              string         `json:"attemptId"`
	EventType              string         `json:"eventType"`
	PreviousStatus         string         `json:"previousStatus,omitempty"`
	NewStatus              string         `json:"newStatus,omitempty"`
	EventPayloadSanitized  map[string]any `json:"eventPayloadSanitized,omitempty"`
	Source                 string         `json:"source,omitempty"`
	RequestID              string         `json:"requestId,omitempty"`
	CorrelationID          string         `json:"correlationId,omitempty"`
	TraceID                string         `json:"traceId,omitempty"`
	OccurredAt             time.Time      `json:"occurredAt"`
}

// ProviderAttemptListResponse is the paginated attempt list response.
type ProviderAttemptListResponse struct {
	Items      []*ProviderAttemptSummary `json:"items"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"pageSize"`
	TotalPages int                       `json:"totalPages"`
}

// MapProviderAttemptSummary converts a model attempt to the summary DTO.
func MapProviderAttemptSummary(a *models.ProviderAttempt) *ProviderAttemptSummary {
	return &ProviderAttemptSummary{
		ID:                  a.ID,
		NotificationID:      a.NotificationID,
		ProviderAccountID:   a.ProviderAccountID,
		TenantID:            a.TenantID,
		Channel:             a.Channel,
		Provider:            a.Provider,
		AttemptNumber:       a.AttemptNumber,
		FallbackSequence:    a.FallbackSequence,
		Status:              a.Status,
		ProviderStatus:      a.ProviderStatus,
		ProviderMessageID:   a.ProviderMessageID,
		RecipientMasked:     a.RecipientMasked,
		ResponseStatusCode:  a.ResponseStatusCode,
		DurationMs:          a.DurationMs,
		Retryable:           a.Retryable,
		NormalizedErrorKind: a.NormalizedErrorKind,
		NormalizedErrorCode: a.NormalizedErrorCode,
		RequestID:           a.RequestID,
		CorrelationID:       a.CorrelationID,
		TraceID:             a.TraceID,
		CreatedAt:           a.CreatedAt,
		CompletedAt:         a.CompletedAt,
	}
}

// MapProviderAttemptDetails converts a model attempt + events to the details DTO.
func MapProviderAttemptDetails(a *models.ProviderAttempt, events []*models.ProviderAttemptEvent) *ProviderAttemptDetails {
	d := &ProviderAttemptDetails{
		ProviderAttemptSummary: *MapProviderAttemptSummary(a),
		ParentAttemptID:        a.ParentAttemptID,
		RequestMethod:          a.RequestMethod,
		RequestURLSanitized:    a.RequestURLSanitized,
		RequestHeadersSanitized: parseStringMap(a.RequestHeadersSanitized),
		RequestBodySanitized:   a.RequestBodySanitized,
		RequestSizeBytes:       a.RequestSizeBytes,
		ResponseHeadersSanitized: parseStringMap(a.ResponseHeadersSanitized),
		ResponseBodySanitized:  a.ResponseBodySanitized,
		ResponseSizeBytes:      a.ResponseSizeBytes,
		BodyTruncated:          a.BodyTruncated,
		OriginalSizeBytes:      a.OriginalSizeBytes,
		CapturedSizeBytes:      a.CapturedSizeBytes,
		ContentHash:            a.ContentHash,
		BodyPreview:            a.BodyPreview,
		ProviderErrorCode:      a.ProviderErrorCode,
		NormalizedErrorMessage: a.NormalizedErrorMessage,
		QueuedAt:               a.QueuedAt,
		StartedAt:              a.StartedAt,
		TimeoutMs:              a.TimeoutMs,
		SpanID:                 a.SpanID,
	}
	if events != nil {
		d.Events = make([]*ProviderAttemptEventResponse, 0, len(events))
		for _, e := range events {
			d.Events = append(d.Events, MapProviderAttemptEvent(e))
		}
	}
	return d
}

func MapProviderAttemptEvent(e *models.ProviderAttemptEvent) *ProviderAttemptEventResponse {
	return &ProviderAttemptEventResponse{
		ID:                    e.ID.String(),
		AttemptID:             e.AttemptID.String(),
		EventType:             e.EventType,
		PreviousStatus:        e.PreviousStatus,
		NewStatus:             e.NewStatus,
		EventPayloadSanitized: parseAnyMap(e.EventPayloadSanitized),
		Source:                e.Source,
		RequestID:             e.RequestID,
		CorrelationID:         e.CorrelationID,
		TraceID:               e.TraceID,
		OccurredAt:            e.OccurredAt,
	}
}

// parseStringMap parses a JSON object stored as text into a string map.
func parseStringMap(raw string) map[string]string {
	if raw == "" {
		return map[string]string{}
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return map[string]string{}
	}
	if len(m) == 0 {
		return map[string]string{}
	}
	return m
}

// parseAnyMap parses a JSON object stored as text into a generic map.
func parseAnyMap(raw string) map[string]any {
	if raw == "" {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return map[string]any{}
	}
	if len(m) == 0 {
		return map[string]any{}
	}
	return m
}
