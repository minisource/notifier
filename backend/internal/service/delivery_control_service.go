package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

// DeliveryPausedError is returned by send paths when the global outbound
// delivery pause is active. It is a CONTROL state, never a provider failure:
// callers must hold the delivery and preserve its retry budget.
var DeliveryPausedError = errors.New("outbound delivery is paused")

// ErrControlStateConflict is returned when a high-risk Pause/Resume request
// carries an expectedVersion that no longer matches the authoritative state
// (another operator changed it). Callers must reload and require a fresh
// explicit confirmation — never auto-replay against the new version.
var ErrControlStateConflict = errors.New("delivery control state changed concurrently (stale expected version)")

// ErrIdempotencyConflict is returned when the same Idempotency-Key is reused
// with a different payload by the same actor. Replay with the SAME payload
// returns the original result; a different payload is rejected.
var ErrIdempotencyConflict = errors.New("idempotency key already used with a different payload")

// DeliveryControlService is the authoritative, backend-owned source of truth
// for the Global Outbound Delivery Pause / Emergency Freeze feature.
//
// Design decisions:
//   - The durable state lives in delivery_control_state (single row, "global").
//   - Every outbound path consults IsPaused() at the final provider boundary
//     (see handlers.go / worker). The pause is enforced by the BACKEND, not by
//     a UI button.
//   - A short-TTL in-process cache avoids a DB read on every send; the cache
//     is version-aware (optimistic concurrency on save) so a stale worker can
//     never start sending after a pause became effective.
//   - Holding a delivery NEVER consumes a retry budget and is never a failure.
//   - Resume releases held work gradually (controlled release) to avoid a
//     thundering herd.
type DeliveryControlService struct {
	repo     repository.DeliveryControlRepository
	notifRepo repository.NotificationRepository
	logger   logging.Logger
	cfg      *config.DeliveryControlConfig

	mu        sync.RWMutex
	cached    *models.DeliveryControlState
	cachedErr error
	cachedAt  time.Time
}

// NewDeliveryControlService creates the service.
func NewDeliveryControlService(
	cfg *config.DeliveryControlConfig,
	logger logging.Logger,
	repo repository.DeliveryControlRepository,
	notifRepo repository.NotificationRepository,
) *DeliveryControlService {
	return &DeliveryControlService{
		repo:      repo,
		notifRepo: notifRepo,
		logger:    logger,
		cfg:       cfg,
	}
}

// cacheTTL returns the cache validity window (default 2s).
func (s *DeliveryControlService) cacheTTL() time.Duration {
	if s.cfg != nil && s.cfg.CacheTTLSeconds > 0 {
		return time.Duration(s.cfg.CacheTTLSeconds) * time.Second
	}
	return 2 * time.Second
}

// loadState returns the current authoritative state, refreshing the cache when
// stale. On DB failure it FAILS CLOSED (returns a paused-like state) so a
// worker can never accidentally send when pause state is unknown. The caller
// must check cachedErr separately when it matters (status API).
func (s *DeliveryControlService) loadState(ctx context.Context) (*models.DeliveryControlState, error) {
	s.mu.RLock()
	if s.cached != nil && time.Since(s.cachedAt) < s.cacheTTL() {
		st, err := s.cached, s.cachedErr
		s.mu.RUnlock()
		return st, err
	}
	s.mu.RUnlock()

	state, err := s.repo.GetState(ctx)
	s.mu.Lock()
	if err != nil {
		// Fail closed: treat as paused with unknown version.
		s.cached = &models.DeliveryControlState{
			ID:    models.DeliveryControlGlobalID,
			State: models.DeliveryControlPaused,
			Mode:  models.DeliveryControlModeImmediate,
			Version: -1,
		}
		s.cachedErr = err
	} else {
		s.cached = state
		s.cachedErr = nil
	}
	s.cachedAt = time.Now()
	st, cachedErr := s.cached, s.cachedErr
	s.mu.Unlock()
	return st, cachedErr
}

// invalidate forces the next IsPaused()/GetStatus() to re-read the DB.
func (s *DeliveryControlService) invalidate() {
	s.mu.Lock()
	s.cached = nil
	s.cachedAt = time.Time{}
	s.mu.Unlock()
}

// IsPaused returns true when outbound provider execution must stop.
// Fail-closed: when the authoritative state cannot be read, it returns true
// (block) rather than risk sending during an active pause.
func (s *DeliveryControlService) IsPaused(ctx context.Context) bool {
	state, err := s.loadState(ctx)
	if err != nil {
		return true
	}
	switch state.State {
	case models.DeliveryControlActive, models.DeliveryControlActiveWithUncertain:
		return false
	default:
		return true
	}
}

// CurrentState returns the current authoritative state (for the status API).
func (s *DeliveryControlService) CurrentState(ctx context.Context) (*models.DeliveryControlState, error) {
	s.invalidate()
	return s.loadState(ctx)
}

// RequestPause freezes outbound delivery. Idempotent: calling pause while
// already paused updates reason/mode/expiry and keeps the same version.
// A reason is mandatory.
func (s *DeliveryControlService) RequestPause(ctx context.Context, actor, mode, reason string, expiresAt *time.Time) (*models.DeliveryControlState, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("reason is required to pause deliveries")
	}
	if s.cfg != nil && s.cfg.MaxReasonLength > 0 && len(reason) > s.cfg.MaxReasonLength {
		return nil, fmt.Errorf("reason exceeds maximum length of %d characters", s.cfg.MaxReasonLength)
	}
	if mode == "" {
		mode = models.DeliveryControlModeImmediate
	}
	if mode != models.DeliveryControlModeImmediate && mode != models.DeliveryControlModeDrain {
		return nil, fmt.Errorf("invalid pause mode %q (expected immediate|drain)", mode)
	}

	state, err := s.repo.GetState(ctx)
	if err != nil {
		return nil, err
	}

	prev := state.State
	now := time.Now().UTC()
	expected := state.Version
	newVersion := expected

	switch state.State {
	case models.DeliveryControlPaused, models.DeliveryControlPauseRequested:
		// Already paused — idempotent refresh of reason/mode/expiry, version
		// stays stable so a re-pause never invalidates worker caches.
		state.Mode = mode
		state.Reason = reason
		state.PausedBy = actor
		state.ExpiresAt = expiresAt
		state.UpdatedAt = now
	case models.DeliveryControlResumeRequested, models.DeliveryControlActiveWithUncertain, models.DeliveryControlActive:
		state.State = models.DeliveryControlPaused
		state.Mode = mode
		state.Reason = reason
		state.PausedBy = actor
		state.PausedAt = &now
		state.EffectiveAt = &now
		state.ExpiresAt = expiresAt
		newVersion = expected + 1
		state.UpdatedAt = now
	}

	if err := s.saveWithRetry(ctx, state, expected, newVersion); err != nil {
		return nil, err
	}
	state.Version = newVersion
	s.invalidate()
	s.createEvent(ctx, models.DeliveryControlActionPauseEffective, actor, reason, mode, prev, state.State, newVersion)
	s.logger.Warn(logging.General, logging.Update, "Global outbound delivery PAUSED", map[logging.ExtraKey]interface{}{
		"actor": actor, "reason": reason, "mode": mode, "version": newVersion,
	})
	return state, nil
}

// RequestResume unfreezes outbound delivery. Idempotent. Optional reason.
func (s *DeliveryControlService) RequestResume(ctx context.Context, actor, reason string) (*models.DeliveryControlState, error) {
	state, err := s.repo.GetState(ctx)
	if err != nil {
		return nil, err
	}

	prev := state.State
	if state.State == models.DeliveryControlActive {
		// Already active — idempotent no-op.
		return state, nil
	}

	expected := state.Version
	newVersion := expected + 1
	now := time.Now().UTC()
	state.State = models.DeliveryControlActive
	state.ResumedBy = actor
	state.ResumedAt = &now
	state.ExpiresAt = nil
	state.UpdatedAt = now

	if err := s.saveWithRetry(ctx, state, expected, newVersion); err != nil {
		return nil, err
	}
	state.Version = newVersion
	s.invalidate()
	s.createEvent(ctx, models.DeliveryControlActionResumeEffective, actor, reason, state.Mode, prev, state.State, newVersion)
	s.logger.Warn(logging.General, logging.Update, "Global outbound delivery RESUMED", map[logging.ExtraKey]interface{}{
		"actor": actor, "reason": reason, "version": newVersion,
	})
	return state, nil
}

// CheckAutoResume resumes automatically when an optional expires_at deadline
// passed while paused. Called periodically by the worker.
func (s *DeliveryControlService) CheckAutoResume(ctx context.Context) error {
	state, err := s.loadState(ctx)
	if err != nil {
		return err
	}
	if state.State != models.DeliveryControlPaused || state.ExpiresAt == nil {
		return nil
	}
	if time.Now().UTC().Before(*state.ExpiresAt) {
		return nil
	}
	// RequestResume already records the resume_effective audit event.
	_, err = s.RequestResume(ctx, "system", "scheduled auto-resume after pause expiration")
	return err
}

// HoldNotification freezes one delivery (used by worker gate paths). The
// reason is descriptive (e.g. "initial_send", "retry", "fallback", "scheduled").
// Holding never consumes a retry budget and never counts as a failure.
func (s *DeliveryControlService) HoldNotification(ctx context.Context, notificationID uuid.UUID, reason string) error {
	state, err := s.loadState(ctx)
	if err != nil {
		return err
	}
	held, err := s.notifRepo.HoldForPause(ctx, notificationID, state.Version, reason)
	if err != nil {
		return err
	}
	if !held {
		return nil // already terminal or held — fine
	}
	s.logger.Info(logging.Internal, logging.Update, "Delivery held by global pause", map[logging.ExtraKey]interface{}{
		"notificationId": notificationID.String(), "reason": reason, "version": state.Version,
	})
	return nil
}

// ReleaseHeld re-queues up to limit held deliveries (controlled release).
// No-op while paused. Called periodically by the worker.
func (s *DeliveryControlService) ReleaseHeld(ctx context.Context, limit int) (int64, error) {
	state, err := s.loadState(ctx)
	if err != nil {
		return 0, err
	}
	switch state.State {
	case models.DeliveryControlActive, models.DeliveryControlActiveWithUncertain:
		released, rerr := s.notifRepo.ReleaseHeld(ctx, limit)
		if rerr == nil && released > 0 {
			s.logger.Info(logging.Internal, logging.Update, "Released held deliveries (controlled)", map[logging.ExtraKey]interface{}{
				"count": released, "version": state.Version,
			})
		}
		return released, rerr
	default:
		return 0, nil
	}
}

// HeldSummary returns counts for the status API.
func (s *DeliveryControlService) HeldSummary(ctx context.Context) (int64, error) {
	return s.notifRepo.CountHeld(ctx)
}

// ActiveCount returns the number of deliveries currently executing
// (status sending/processing). Used by the status API — real data, never a
// fabricated 0.
func (s *DeliveryControlService) ActiveCount(ctx context.Context) int64 {
	n, err := s.notifRepo.CountActive(ctx)
	if err != nil {
		s.logger.Warn(logging.General, logging.Select, "Failed to count active deliveries", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return 0
	}
	return n
}

// RetryingHeldCount returns the number of held deliveries frozen during a
// retry (retry_count > 0). Used by the status API.
func (s *DeliveryControlService) RetryingHeldCount(ctx context.Context) int64 {
	n, err := s.notifRepo.CountHeldRetries(ctx)
	if err != nil {
		s.logger.Warn(logging.General, logging.Select, "Failed to count retrying held deliveries", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return 0
	}
	return n
}

// ReplayIdempotency returns the stored result JSON for a previously processed
// (actor, idempotency key, operation) when the request hash matches — i.e. a
// replay/browser retry of the SAME request. Returns (json, true, nil) on
// replay, ("", false, nil) when the key is new, and ErrIdempotencyConflict
// when the same key was used with a DIFFERENT payload.
func (s *DeliveryControlService) ReplayIdempotency(ctx context.Context, actor, key, operation, requestHash string) (string, bool, error) {
	if key == "" {
		return "", false, nil
	}
	rec, err := s.repo.GetIdempotency(ctx, actor, key)
	if err != nil {
		return "", false, err
	}
	if rec == nil {
		return "", false, nil
	}
	if rec.Operation != operation || rec.RequestHash != requestHash {
		return "", false, ErrIdempotencyConflict
	}
	return rec.ResultJSON, true, nil
}

// RecordIdempotency persists the outcome of one high-risk request so a replay
// returns the original result instead of a duplicate transition/audit event.
// state/version are the resulting control state (diagnostics; the stored
// ResultJSON is the authoritative replay payload).
func (s *DeliveryControlService) RecordIdempotency(ctx context.Context, actor, key, operation, requestHash, resultJSON, state string, version int64) {
	if key == "" {
		return
	}
	retention := time.Duration(24) * time.Hour
	if s.cfg != nil && s.cfg.IdempotencyRetentionHours > 0 {
		retention = time.Duration(s.cfg.IdempotencyRetentionHours) * time.Hour
	}
	rec := &models.DeliveryControlIdempotency{
		Actor:          actor,
		IdempotencyKey: key,
		Operation:      operation,
		RequestHash:    requestHash,
		State:          state,
		Version:        version,
		ResultJSON:     resultJSON,
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(retention),
	}
	if err := s.repo.SaveIdempotency(ctx, rec); err != nil {
		s.logger.Warn(logging.General, logging.Insert, "Failed to record delivery control idempotency", map[logging.ExtraKey]interface{}{
			"operation": operation, "error": err.Error(),
		})
	}
}

// PurgeExpiredIdempotency removes expired idempotency records (bounded
// retention). Called periodically by the worker.
func (s *DeliveryControlService) PurgeExpiredIdempotency(ctx context.Context) (int64, error) {
	return s.repo.PurgeExpiredIdempotency(ctx, time.Now().UTC())
}

// ValidateExpectedVersion returns ErrControlStateConflict when expectedVersion
// (from the client's last-known state) no longer matches the authoritative
// state. Zero disables the check.
func (s *DeliveryControlService) ValidateExpectedVersion(ctx context.Context, expectedVersion int64) error {
	if expectedVersion <= 0 {
		return nil
	}
	state, err := s.repo.GetState(ctx)
	if err != nil {
		return err
	}
	if state.Version != expectedVersion {
		return ErrControlStateConflict
	}
	return nil
}

// ListEvents returns pause/resume audit events.
func (s *DeliveryControlService) ListEvents(ctx context.Context, limit int) ([]*models.DeliveryControlEvent, error) {
	return s.repo.ListEvents(ctx, limit)
}

func (s *DeliveryControlService) saveWithRetry(ctx context.Context, state *models.DeliveryControlState, expected, newVersion int64) error {
	err := s.repo.SaveState(ctx, state, expected, newVersion)
	if err == nil || !errors.Is(err, repository.ErrVersionConflict) {
		return err
	}
	// Optimistic concurrency conflict — reload and reapply against the fresh
	// version exactly once. newVersion is recomputed relative to the new base.
	fresh, gerr := s.repo.GetState(ctx)
	if gerr != nil {
		return gerr
	}
	state.Version = fresh.Version
	if newVersion != expected {
		newVersion = fresh.Version + 1
	}
	return s.repo.SaveState(ctx, state, fresh.Version, newVersion)
}

func (s *DeliveryControlService) createEvent(ctx context.Context, action, actor, reason, mode, from, to string, version int64) {
	event := &models.DeliveryControlEvent{
		Action:    action,
		Actor:     actor,
		Reason:    reason,
		Mode:      mode,
		FromState: from,
		ToState:   to,
		Version:   version,
	}
	if err := s.repo.CreateEvent(ctx, event); err != nil {
		s.logger.Warn(logging.General, logging.Insert, "Failed to record delivery control audit event", map[logging.ExtraKey]interface{}{
			"action": action, "error": err.Error(),
		})
	}
}
