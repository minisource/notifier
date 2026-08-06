package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

// nopLogger is a discard logger for tests.
type nopLogger struct{}

func (nopLogger) Init()                                                                                              {}
func (nopLogger) Debug(cat logging.Category, sub logging.SubCategory, msg string, extra map[logging.ExtraKey]interface{}) {}
func (nopLogger) Debugf(template string, args ...interface{})                                                       {}
func (nopLogger) Info(cat logging.Category, sub logging.SubCategory, msg string, extra map[logging.ExtraKey]interface{})  {}
func (nopLogger) Infof(template string, args ...interface{})                                                        {}
func (nopLogger) Warn(cat logging.Category, sub logging.SubCategory, msg string, extra map[logging.ExtraKey]interface{})  {}
func (nopLogger) Warnf(template string, args ...interface{})                                                        {}
func (nopLogger) Error(cat logging.Category, sub logging.SubCategory, msg string, extra map[logging.ExtraKey]interface{}) {}
func (nopLogger) Errorf(template string, args ...interface{})                                                       {}
func (nopLogger) Fatal(cat logging.Category, sub logging.SubCategory, msg string, extra map[logging.ExtraKey]interface{}) {}
func (nopLogger) Fatalf(template string, args ...interface{})                                                       {}

// ---- fakes ----

type fakeDeliveryControlRepo struct {
	mu          sync.Mutex
	state       *models.DeliveryControlState
	events      []*models.DeliveryControlEvent
	idempotency map[string]*models.DeliveryControlIdempotency
	err         error
}

func (f *fakeDeliveryControlRepo) GetState(ctx context.Context) (*models.DeliveryControlState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	if f.state == nil {
		now := time.Now().UTC()
		f.state = &models.DeliveryControlState{
			ID:        models.DeliveryControlGlobalID,
			State:     models.DeliveryControlActive,
			Mode:      models.DeliveryControlModeImmediate,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	cp := *f.state
	return &cp, nil
}

func (f *fakeDeliveryControlRepo) SaveState(ctx context.Context, state *models.DeliveryControlState, expectedVersion, newVersion int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	if f.state != nil && f.state.Version != expectedVersion {
		return repository.ErrVersionConflict
	}
	cp := *state
	cp.Version = newVersion
	f.state = &cp
	return nil
}

func (f *fakeDeliveryControlRepo) CreateEvent(ctx context.Context, event *models.DeliveryControlEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, event)
	return nil
}

func (f *fakeDeliveryControlRepo) GetIdempotency(ctx context.Context, actor, key string) (*models.DeliveryControlIdempotency, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.idempotency == nil {
		return nil, nil
	}
	rec, ok := f.idempotency[actor+"|"+key]
	if !ok || rec.ExpiresAt.Before(time.Now().UTC()) {
		return nil, nil
	}
	cp := *rec
	return &cp, nil
}

func (f *fakeDeliveryControlRepo) SaveIdempotency(ctx context.Context, rec *models.DeliveryControlIdempotency) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.idempotency == nil {
		f.idempotency = map[string]*models.DeliveryControlIdempotency{}
	}
	cp := *rec
	f.idempotency[rec.Actor+"|"+rec.IdempotencyKey] = &cp
	return nil
}

func (f *fakeDeliveryControlRepo) PurgeExpiredIdempotency(ctx context.Context, before time.Time) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var removed int64
	for k, rec := range f.idempotency {
		if rec.ExpiresAt.Before(before) {
			delete(f.idempotency, k)
			removed++
		}
	}
	return removed, nil
}

func (f *fakeDeliveryControlRepo) ListEvents(ctx context.Context, limit int) ([]*models.DeliveryControlEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*models.DeliveryControlEvent, 0, len(f.events))
	for i := len(f.events) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, f.events[i])
	}
	return out, nil
}

// fakeNotifRepo embeds the full interface (nil) and overrides only the
// delivery-pause methods the service uses.
type fakeNotifRepo struct {
	repository.NotificationRepository

	mu    sync.Mutex
	held  map[uuid.UUID]bool
	released int64
}

func (f *fakeNotifRepo) HoldForPause(ctx context.Context, id uuid.UUID, version int64, reason string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.held == nil {
		f.held = map[uuid.UUID]bool{}
	}
	f.held[id] = true
	return true, nil
}

func (f *fakeNotifRepo) CountHeld(ctx context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return int64(len(f.held)), nil
}

func (f *fakeNotifRepo) ReleaseHeld(ctx context.Context, limit int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.released += int64(limit)
	return int64(limit), nil
}

func (f *fakeNotifRepo) ListHeld(ctx context.Context, limit, offset int) ([]*models.Notification, int64, error) {
	return nil, 0, nil
}

func (f *fakeNotifRepo) RecoverStuckSending(ctx context.Context, cutoff time.Time) (int64, error) {
	return 0, nil
}

func newTestDeliveryControl(fake *fakeDeliveryControlRepo, notif *fakeNotifRepo) *DeliveryControlService {
	cfg := &config.DeliveryControlConfig{Enabled: true, CacheTTLSeconds: 60}
	return NewDeliveryControlService(cfg, nopLogger{}, fake, notif)
}

// ---- tests ----

func TestDeliveryControl_InitiallyActive(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	ctx := context.Background()
	if dc.IsPaused(ctx) {
		t.Fatal("expected not paused initially")
	}
}

func TestDeliveryControl_PauseRequiresReason(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	_, err := dc.RequestPause(context.Background(), "admin", "immediate", "", nil)
	if err == nil {
		t.Fatal("expected error when reason is empty")
	}
}

func TestDeliveryControl_PauseThenIsPaused(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	ctx := context.Background()
	state, err := dc.RequestPause(ctx, "admin", "immediate", "emergency", nil)
	if err != nil {
		t.Fatalf("pause failed: %v", err)
	}
	if state.State != models.DeliveryControlPaused {
		t.Fatalf("expected paused, got %s", state.State)
	}
	if !dc.IsPaused(ctx) {
		t.Fatal("expected IsPaused true after pause")
	}
	if state.Version != 2 {
		t.Fatalf("expected version bumped to 2, got %d", state.Version)
	}
}

func TestDeliveryControl_PauseIdempotent_NoVersionBump(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	ctx := context.Background()
	s1, err := dc.RequestPause(ctx, "admin", "immediate", "emergency", nil)
	if err != nil {
		t.Fatalf("first pause failed: %v", err)
	}
	s2, err := dc.RequestPause(ctx, "admin", "immediate", "still emergency", nil)
	if err != nil {
		t.Fatalf("second pause failed: %v", err)
	}
	if s1.Version != s2.Version {
		t.Fatalf("idempotent pause must not bump version: %d vs %d", s1.Version, s2.Version)
	}
	if s2.Reason != "still emergency" {
		t.Fatalf("expected reason updated, got %q", s2.Reason)
	}
}

func TestDeliveryControl_ResumeThenActive(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	ctx := context.Background()
	if _, err := dc.RequestPause(ctx, "admin", "immediate", "emergency", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := dc.RequestResume(ctx, "admin", "all clear"); err != nil {
		t.Fatalf("resume failed: %v", err)
	}
	if dc.IsPaused(ctx) {
		t.Fatal("expected not paused after resume")
	}
	state, _ := dc.CurrentState(ctx)
	if state.State != models.DeliveryControlActive {
		t.Fatalf("expected active after resume, got %s", state.State)
	}
}

func TestDeliveryControl_ResumeIdempotent(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	ctx := context.Background()
	if _, err := dc.RequestResume(ctx, "admin", ""); err != nil {
		t.Fatalf("resume when already active failed: %v", err)
	}
	if dc.IsPaused(ctx) {
		t.Fatal("expected active after idempotent resume")
	}
}

func TestDeliveryControl_InvalidModeRejected(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	_, err := dc.RequestPause(context.Background(), "admin", "instant", "x", nil)
	if err == nil {
		t.Fatal("expected invalid mode rejected")
	}
}

func TestDeliveryControl_AutoResumeAfterExpiry(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	ctx := context.Background()
	past := time.Now().UTC().Add(-time.Minute)
	if _, err := dc.RequestPause(ctx, "admin", "immediate", "maintenance", &past); err != nil {
		t.Fatal(err)
	}
	if !dc.IsPaused(ctx) {
		t.Fatal("expected paused while deadline not yet checked")
	}
	// Trigger the auto-resume check (simulates the periodic worker tick).
	if err := dc.CheckAutoResume(ctx); err != nil {
		t.Fatalf("auto resume failed: %v", err)
	}
	dc.invalidate()
	if dc.IsPaused(ctx) {
		t.Fatal("expected active after auto-resume")
	}
}

func TestDeliveryControl_AutoResumeNotBeforeDeadline(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	ctx := context.Background()
	future := time.Now().UTC().Add(time.Hour)
	if _, err := dc.RequestPause(ctx, "admin", "immediate", "maintenance", &future); err != nil {
		t.Fatal(err)
	}
	if err := dc.CheckAutoResume(ctx); err != nil {
		t.Fatalf("check failed: %v", err)
	}
	if !dc.IsPaused(ctx) {
		t.Fatal("expected still paused before deadline")
	}
}

func TestDeliveryControl_FailClosedOnRepoError(t *testing.T) {
	fake := &fakeDeliveryControlRepo{err: errors.New("db down")}
	dc := newTestDeliveryControl(fake, &fakeNotifRepo{})
	if !dc.IsPaused(context.Background()) {
		t.Fatal("expected fail-closed (paused) when state cannot be read")
	}
}

func TestDeliveryControl_HoldNotification(t *testing.T) {
	notif := &fakeNotifRepo{}
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, notif)
	ctx := context.Background()
	id := uuid.New()
	if err := dc.HoldNotification(ctx, id, "retry"); err != nil {
		t.Fatalf("hold failed: %v", err)
	}
	if len(notif.held) != 1 || !notif.held[id] {
		t.Fatal("expected notification held")
	}
	count, _ := dc.HeldSummary(ctx)
	if count != 1 {
		t.Fatalf("expected 1 held, got %d", count)
	}
}

func TestDeliveryControl_ReasonLengthRejected(t *testing.T) {
	cfg := &config.DeliveryControlConfig{Enabled: true, CacheTTLSeconds: 60, MaxReasonLength: 10}
	dc := NewDeliveryControlService(cfg, nopLogger{}, &fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	_, err := dc.RequestPause(context.Background(), "admin", "immediate", "this reason is way too long", nil)
	if err == nil {
		t.Fatal("expected reason-length rejection")
	}
	if _, err := dc.RequestPause(context.Background(), "admin", "immediate", "short", nil); err != nil {
		t.Fatalf("short reason must be accepted: %v", err)
	}
}

func TestDeliveryControl_ValidateExpectedVersion(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	ctx := context.Background()

	// Fresh state version is 1 — matching expectedVersion is accepted.
	if err := dc.ValidateExpectedVersion(ctx, 1); err != nil {
		t.Fatalf("matching version must pass: %v", err)
	}
	// Stale expectedVersion is rejected with the typed conflict error.
	if err := dc.ValidateExpectedVersion(ctx, 1); err != nil {
		t.Fatalf("expected pass: %v", err)
	}
	// Bump the version via a pause, then a stale expected version must fail.
	if _, err := dc.RequestPause(ctx, "admin", "immediate", "hold", nil); err != nil {
		t.Fatal(err)
	}
	dc.invalidate()
	if err := dc.ValidateExpectedVersion(ctx, 1); !errors.Is(err, ErrControlStateConflict) {
		t.Fatalf("expected ErrControlStateConflict, got %v", err)
	}
	if err := dc.ValidateExpectedVersion(ctx, 2); err != nil {
		t.Fatalf("current version must pass: %v", err)
	}
	// Zero disables the check.
	if err := dc.ValidateExpectedVersion(ctx, 0); err != nil {
		t.Fatalf("zero must disable check: %v", err)
	}
}

func TestDeliveryControl_IdempotencyReplayReturnsOriginal(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	ctx := context.Background()

	// No record yet — not a replay.
	if _, replay, err := dc.ReplayIdempotency(ctx, "admin", "key-1", "pause", "hash-A"); err != nil || replay {
		t.Fatalf("expected no replay for fresh key (err=%v replay=%v)", err, replay)
	}

	// Record a processed request.
	dc.RecordIdempotency(ctx, "admin", "key-1", "pause", "hash-A", `{"state":"paused"}`, "paused", 2)

	// Same actor + same key + same hash -> replay with stored result.
	result, replay, err := dc.ReplayIdempotency(ctx, "admin", "key-1", "pause", "hash-A")
	if err != nil || !replay {
		t.Fatalf("expected replay (err=%v replay=%v)", err, replay)
	}
	if result != `{"state":"paused"}` {
		t.Fatalf("expected stored result, got %q", result)
	}

	// Same key + different payload -> conflict.
	if _, _, err := dc.ReplayIdempotency(ctx, "admin", "key-1", "pause", "hash-B"); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected ErrIdempotencyConflict, got %v", err)
	}
	// Same key + different operation -> conflict.
	if _, _, err := dc.ReplayIdempotency(ctx, "admin", "key-1", "resume", "hash-A"); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected ErrIdempotencyConflict for op mismatch, got %v", err)
	}
	// Different actor with same key is isolated (treated as new).
	if _, replay, err := dc.ReplayIdempotency(ctx, "other-admin", "key-1", "pause", "hash-A"); err != nil || replay {
		t.Fatalf("expected isolation per actor (err=%v replay=%v)", err, replay)
	}
}

func TestDeliveryControl_PurgeExpiredIdempotency(t *testing.T) {
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, &fakeNotifRepo{})
	ctx := context.Background()
	dc.RecordIdempotency(ctx, "admin", "expiring", "pause", "h", `{}`, "paused", 2)
	if _, replay, _ := dc.ReplayIdempotency(ctx, "admin", "expiring", "pause", "h"); !replay {
		t.Fatal("expected record present before purge")
	}
	repo := dc.repo.(*fakeDeliveryControlRepo)
	repo.mu.Lock()
	for _, rec := range repo.idempotency {
		rec.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	}
	repo.mu.Unlock()
	if n, err := dc.PurgeExpiredIdempotency(ctx); err != nil || n != 1 {
		t.Fatalf("expected 1 purged, got %d (err=%v)", n, err)
	}
	if _, replay, _ := dc.ReplayIdempotency(ctx, "admin", "expiring", "pause", "h"); replay {
		t.Fatal("expected record gone after purge")
	}
}

func TestDeliveryControl_ReleaseOnlyWhenActive(t *testing.T) {
	notif := &fakeNotifRepo{}
	dc := newTestDeliveryControl(&fakeDeliveryControlRepo{}, notif)
	ctx := context.Background()

	// While paused, release must be a no-op.
	if _, err := dc.RequestPause(ctx, "admin", "immediate", "hold", nil); err != nil {
		t.Fatal(err)
	}
	released, err := dc.ReleaseHeld(ctx, 50)
	if err != nil {
		t.Fatal(err)
	}
	if released != 0 {
		t.Fatalf("expected no release while paused, got %d", released)
	}

	// After resume, release re-queues a bounded batch.
	if _, err := dc.RequestResume(ctx, "admin", "go"); err != nil {
		t.Fatal(err)
	}
	released, err = dc.ReleaseHeld(ctx, 25)
	if err != nil {
		t.Fatal(err)
	}
	if released != 25 {
		t.Fatalf("expected 25 released, got %d", released)
	}
}
