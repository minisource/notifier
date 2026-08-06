package repository

import (
	"context"
	"time"

	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DeliveryControlRepository persists the authoritative global outbound
// delivery pause state and its audit events. The state is a single row
// (ID = models.DeliveryControlGlobalID); the Version field is a monotonic
// generation used to invalidate worker caches and reject stale control
// messages.
type DeliveryControlRepository interface {
	// GetState returns the current control state, creating the default
	// "active" row on first use (idempotent).
	GetState(ctx context.Context) (*models.DeliveryControlState, error)
	// SaveState persists the state with optimistic concurrency: it only
	// overwrites the row if the stored version still equals expectedVersion,
	// then writes newVersion. Returns ErrVersionConflict so callers can
	// reload and retry. Idempotent refreshes pass newVersion == expectedVersion.
	SaveState(ctx context.Context, state *models.DeliveryControlState, expectedVersion, newVersion int64) error
	// CreateEvent appends an audit event.
	CreateEvent(ctx context.Context, event *models.DeliveryControlEvent) error
	// ListEvents returns audit events newest-first, up to limit.
	ListEvents(ctx context.Context, limit int) ([]*models.DeliveryControlEvent, error)

	// GetIdempotency returns a previously processed pause/resume request for
	// (actor, key), or nil when absent/expired. Used for replay protection.
	GetIdempotency(ctx context.Context, actor, key string) (*models.DeliveryControlIdempotency, error)
	// SaveIdempotency records a processed request so replays return the
	// original result. Unique on (actor, idempotency_key).
	SaveIdempotency(ctx context.Context, rec *models.DeliveryControlIdempotency) error
	// PurgeExpiredIdempotency deletes idempotency rows whose ExpiresAt is in
	// the past (bounded retention), returning the number removed.
	PurgeExpiredIdempotency(ctx context.Context, before time.Time) (int64, error)
}

type deliveryControlRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

// NewDeliveryControlRepository creates the repository.
func NewDeliveryControlRepository(db *gorm.DB, logger logging.Logger) DeliveryControlRepository {
	return &deliveryControlRepository{db: db, logger: logger}
}

func (r *deliveryControlRepository) GetState(ctx context.Context) (*models.DeliveryControlState, error) {
	var state models.DeliveryControlState
	err := r.db.WithContext(ctx).Where("id = ?", models.DeliveryControlGlobalID).First(&state).Error
	if err == nil {
		return &state, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// First use — create the default active row.
	now := time.Now().UTC()
	state = models.DeliveryControlState{
		ID:        models.DeliveryControlGlobalID,
		State:     models.DeliveryControlActive,
		Mode:      models.DeliveryControlModeImmediate,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.db.WithContext(ctx).Create(&state).Error; err != nil {
		// Race with another instance: read again instead of failing.
		var existing models.DeliveryControlState
		if err2 := r.db.WithContext(ctx).Where("id = ?", models.DeliveryControlGlobalID).First(&existing).Error; err2 == nil {
			return &existing, nil
		}
		return nil, err
	}
	return &state, nil
}

func (r *deliveryControlRepository) SaveState(ctx context.Context, state *models.DeliveryControlState, expectedVersion, newVersion int64) error {
	if state == nil {
		return nil
	}
	state.UpdatedAt = time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&models.DeliveryControlState{}).
		Where("id = ? AND version = ?", state.ID, expectedVersion).
		Updates(map[string]interface{}{
			"state":         state.State,
			"mode":          state.Mode,
			"reason":        state.Reason,
			"paused_by":     state.PausedBy,
			"paused_at":     state.PausedAt,
			"effective_at":  state.EffectiveAt,
			"expires_at":    state.ExpiresAt,
			"resumed_by":    state.ResumedBy,
			"resumed_at":    state.ResumedAt,
			"version":       newVersion,
			"updated_at":    state.UpdatedAt,
		})
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, "Failed to save delivery control state", map[logging.ExtraKey]interface{}{
			"error": result.Error.Error(),
		})
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrVersionConflict
	}
	return nil
}

func (r *deliveryControlRepository) CreateEvent(ctx context.Context, event *models.DeliveryControlEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *deliveryControlRepository) GetIdempotency(ctx context.Context, actor, key string) (*models.DeliveryControlIdempotency, error) {
	if key == "" {
		return nil, nil
	}
	var rec models.DeliveryControlIdempotency
	err := r.db.WithContext(ctx).
		Where("actor = ? AND idempotency_key = ?", actor, key).
		First(&rec).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	if rec.ExpiresAt.Before(time.Now().UTC()) {
		return nil, nil // expired — treat as a new request
	}
	return &rec, nil
}

func (r *deliveryControlRepository) SaveIdempotency(ctx context.Context, rec *models.DeliveryControlIdempotency) error {
	// The (actor, idempotency_key) composite is UNIQUE. Two concurrent
	// identical requests may both pass the replay check before either writes;
	// the duplicate insert is then a replay, NOT a second transition, so it is
	// safely ignored (ON CONFLICT DO NOTHING).
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(rec).Error
}

func (r *deliveryControlRepository) PurgeExpiredIdempotency(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", before).
		Delete(&models.DeliveryControlIdempotency{})
	return result.RowsAffected, result.Error
}

func (r *deliveryControlRepository) ListEvents(ctx context.Context, limit int) ([]*models.DeliveryControlEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var out []*models.DeliveryControlEvent
	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}
