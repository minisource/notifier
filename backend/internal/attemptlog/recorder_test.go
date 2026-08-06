package attemptlog

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/minisource/notifier/internal/models"
)

// fakeRepo is an in-memory AttemptRepository for recorder tests.
type fakeRepo struct {
	attempts []*models.ProviderAttempt
	events   []*models.ProviderAttemptEvent
}

func (f *fakeRepo) CreateAttempt(_ context.Context, a *models.ProviderAttempt) error {
	a.ID = uuid.New()
	f.attempts = append(f.attempts, a)
	return nil
}
func (f *fakeRepo) UpdateAttempt(_ context.Context, a *models.ProviderAttempt) error { return nil }
func (f *fakeRepo) AddEvent(_ context.Context, e *models.ProviderAttemptEvent) error {
	f.events = append(f.events, e)
	return nil
}

func TestRecorderDisabled_NoPersistence(t *testing.T) {
	repo := &fakeRepo{}
	rec := NewRecorder(repo, nil, DefaultRedactionOptions(), false)
	att := rec.Start(context.Background(), StartInput{NotificationID: uuid.New(), Channel: "sms", Provider: "mock", AttemptNumber: 1})
	if att != nil {
		t.Fatalf("recorder disabled should return nil attempt")
	}
	if len(repo.attempts) != 0 {
		t.Fatalf("no persistence expected when disabled")
	}
}

func TestRecorderStart_FinishesLifecycle(t *testing.T) {
	repo := &fakeRepo{}
	rec := NewRecorder(repo, nil, DefaultRedactionOptions(), true)
	ctx := context.Background()
	nid := uuid.New()

	att := rec.Start(ctx, StartInput{
		NotificationID:      nid,
		Channel:             "sms",
		Provider:            "mock",
		AttemptNumber:       1,
		RequestMethod:       "POST",
		RequestURLSanitized: "sms://mock/send",
		RequestBodySanitized: `{"template":"verify"}`,
		ContentHash:         "abc",
		BodyPreview:         "سلام",
		CorrelationID:       nid.String(),
	})
	if att == nil {
		t.Fatalf("expected attempt to be created")
	}
	if att.Status != string(StatusQueued) {
		t.Fatalf("expected queued status, got %q", att.Status)
	}
	if att.RequestBodySanitized == "" {
		t.Fatalf("request body should be persisted")
	}
	if att.RequestID == "" {
		t.Fatalf("request id should be auto-generated")
	}

	rec.MarkSending(ctx, att)
	if att.Status != string(StatusSending) {
		t.Fatalf("expected sending status, got %q", att.Status)
	}
	if att.StartedAt == nil {
		t.Fatalf("startedAt should be set")
	}

	rec.Finish(ctx, att, FinishInput{
		Status:            StatusAccepted,
		ProviderStatus:    "accepted",
		ProviderMessageID: "msg-1",
		StartedAt:         *att.StartedAt,
	})
	if att.Status != string(StatusAccepted) {
		t.Fatalf("expected accepted status, got %q", att.Status)
	}
	if att.ProviderMessageID != "msg-1" {
		t.Fatalf("provider message id not persisted")
	}
	if att.CompletedAt == nil {
		t.Fatalf("completedAt should be set")
	}
	if att.DurationMs < 0 {
		t.Fatalf("duration should be >= 0, got %d", att.DurationMs)
	}

	// Events: attempt_created + request_started + provider_accepted
	if len(repo.events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(repo.events))
	}
	if repo.events[0].EventType != "attempt_created" {
		t.Fatalf("first event should be attempt_created, got %q", repo.events[0].EventType)
	}
	if repo.events[1].EventType != "request_started" {
		t.Fatalf("second event should be request_started, got %q", repo.events[1].EventType)
	}
	if repo.events[2].EventType != "provider_accepted" {
		t.Fatalf("third event should be provider_accepted, got %q", repo.events[2].EventType)
	}
}

func TestRecorderFailure_FinalizesFailed(t *testing.T) {
	repo := &fakeRepo{}
	rec := NewRecorder(repo, nil, DefaultRedactionOptions(), true)
	ctx := context.Background()

	att := rec.Start(ctx, StartInput{NotificationID: uuid.New(), Channel: "email", Provider: "smtp", AttemptNumber: 1})
	rec.MarkSending(ctx, att)
	rec.Finish(ctx, att, FinishInput{
		Status:                 StatusFailed,
		NormalizedErrorKind:    string(ErrKindNetwork),
		NormalizedErrorCode:    "network_error",
		NormalizedErrorMessage: "dial tcp: connection refused",
		Retryable:              true,
		StartedAt:              *att.StartedAt,
	})

	if att.Status != string(StatusFailed) {
		t.Fatalf("expected failed status")
	}
	if !att.Retryable {
		t.Fatalf("expected retryable true")
	}
	if att.NormalizedErrorCode != "network_error" {
		t.Fatalf("normalized error code not persisted")
	}
	if len(repo.events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(repo.events))
	}
	if repo.events[2].EventType != "request_failed" {
		t.Fatalf("terminal event should be request_failed, got %q", repo.events[2].EventType)
	}
}
