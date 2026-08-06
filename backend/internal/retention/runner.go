package retention

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-common/retention"
	"gorm.io/gorm"
)

// NotifierRunner implements safe, cursor-based batch deletion for Notifier log tables.
type NotifierRunner struct {
	db     *gorm.DB
	logger logging.Logger
}

func NewNotifierRunner(db *gorm.DB, logger logging.Logger) *NotifierRunner {
	return &NotifierRunner{db: db, logger: logger}
}

// ComputeCountCutoff queries the DB for the created_at timestamp of the
// Nth newest record (where N = keepLatest). Records older than this
// threshold are eligible for count-based cleanup.
// Returns a zero time if keepLatest <= 0 or no records exist.
func (r *NotifierRunner) ComputeCountCutoff(ctx context.Context, category string, keepLatest int) (time.Time, error) {
	if keepLatest <= 0 {
		return time.Time{}, nil
	}
	var row struct{ CreatedAt time.Time }
	err := r.db.WithContext(ctx).
		Table(category).
		Select("created_at").
		Order("created_at DESC, id DESC").
		Offset(keepLatest - 1).
		Limit(1).
		Scan(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return time.Time{}, nil // fewer than N records exist
		}
		return time.Time{}, fmt.Errorf("compute count cutoff for %s: %w", category, err)
	}
	return row.CreatedAt, nil
}

func (r *NotifierRunner) NewSharedRunner(snapshot retention.RunSnapshot) (*retention.BatchRunner, error) {
	switch snapshot.Category {
	case CategoryNotificationLogs.String():
		return retention.NewBatchRunner(snapshot, r.notifLogsEligibility, r.notifLogsDelete), nil
	case CategoryProviderAttempts.String():
		return retention.NewBatchRunner(snapshot, r.attemptsEligibility, r.attemptsDelete), nil
	case CategoryProviderBalanceSnapshots.String():
		return retention.NewBatchRunner(snapshot, r.balanceEligibility, r.balanceDelete), nil
	default:
		return nil, fmt.Errorf("%w: %s", retention.ErrCategoryProtected, snapshot.Category)
	}
}

// ── notification_logs ────────────────────────────────────────────────

func (r *NotifierRunner) notifLogsEligibility(ctx context.Context, snapshot retention.RunSnapshot) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("notification_logs").
		Where("created_at < ?", snapshot.Cutoff).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("notification_logs eligibility: %w", err)
	}
	return count, nil
}

func (r *NotifierRunner) notifLogsDelete(ctx context.Context, snapshot retention.RunSnapshot, lastCreatedAt time.Time, lastID uuid.UUID) (int64, time.Time, uuid.UUID, bool, error) {
	type row struct {
		ID        uuid.UUID
		CreatedAt time.Time
	}
	var ids []row
	q := r.db.WithContext(ctx).Table("notification_logs").
		Select("id, created_at").
		Where("created_at < ?", snapshot.Cutoff).
		Order("created_at ASC, id ASC").
		Limit(snapshot.BatchSize)
	if lastID != uuid.Nil {
		q = q.Where("(created_at, id) > (?, ?)", lastCreatedAt, lastID.String())
	}
	if err := q.Find(&ids).Error; err != nil {
		return 0, time.Time{}, uuid.Nil, false, fmt.Errorf("notification_logs cursor: %w", err)
	}
	if len(ids) == 0 {
		return 0, time.Time{}, uuid.Nil, false, nil
	}
	idList := make([]uuid.UUID, len(ids))
	for i, row := range ids {
		idList[i] = row.ID
	}
	res := r.db.WithContext(ctx).Table("notification_logs").Where("id IN ?", idList).Delete(nil)
	if res.Error != nil {
		return 0, time.Time{}, uuid.Nil, false, fmt.Errorf("notification_logs delete: %w", res.Error)
	}
	last := ids[len(ids)-1]
	return res.RowsAffected, last.CreatedAt, last.ID, len(ids) == snapshot.BatchSize, nil
}

// ── notification_provider_attempts + events ──────────────────────────

func (r *NotifierRunner) attemptsEligibility(ctx context.Context, snapshot retention.RunSnapshot) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("notification_provider_attempts").
		Where("created_at < ?", snapshot.Cutoff).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("provider_attempts eligibility: %w", err)
	}
	return count, nil
}

func (r *NotifierRunner) attemptsDelete(ctx context.Context, snapshot retention.RunSnapshot, lastCreatedAt time.Time, lastID uuid.UUID) (int64, time.Time, uuid.UUID, bool, error) {
	type row struct {
		ID        uuid.UUID
		CreatedAt time.Time
	}
	var ids []row
	q := r.db.WithContext(ctx).Table("notification_provider_attempts").
		Select("id, created_at").
		Where("created_at < ?", snapshot.Cutoff).
		Order("created_at ASC, id ASC").
		Limit(snapshot.BatchSize)
	if lastID != uuid.Nil {
		q = q.Where("(created_at, id) > (?, ?)", lastCreatedAt, lastID.String())
	}
	if err := q.Find(&ids).Error; err != nil {
		return 0, time.Time{}, uuid.Nil, false, fmt.Errorf("provider_attempts cursor: %w", err)
	}
	if len(ids) == 0 {
		return 0, time.Time{}, uuid.Nil, false, nil
	}
	idList := make([]uuid.UUID, len(ids))
	for i, row := range ids {
		idList[i] = row.ID
	}
	// Delete events first (FK safety), then attempt rows
	if err := r.db.WithContext(ctx).Table("notification_provider_attempt_events").
		Where("attempt_id IN ?", idList).Delete(nil).Error; err != nil {
		return 0, time.Time{}, uuid.Nil, false, fmt.Errorf("attempt_events delete: %w", err)
	}
	// Unscoped hard-delete because ProviderAttempt has gorm.DeletedAt (soft delete)
	res := r.db.WithContext(ctx).Unscoped().Table("notification_provider_attempts").
		Where("id IN ?", idList).Delete(nil)
	if res.Error != nil {
		return 0, time.Time{}, uuid.Nil, false, fmt.Errorf("provider_attempts delete: %w", res.Error)
	}
	last := ids[len(ids)-1]
	return res.RowsAffected, last.CreatedAt, last.ID, len(ids) == snapshot.BatchSize, nil
}

// ── provider_balance_snapshots ───────────────────────────────────────

func (r *NotifierRunner) balanceEligibility(ctx context.Context, snapshot retention.RunSnapshot) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("provider_balance_snapshots").
		Where("created_at < ?", snapshot.Cutoff).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("balance_snapshots eligibility: %w", err)
	}
	return count, nil
}

func (r *NotifierRunner) balanceDelete(ctx context.Context, snapshot retention.RunSnapshot, lastCreatedAt time.Time, lastID uuid.UUID) (int64, time.Time, uuid.UUID, bool, error) {
	type row struct {
		ID        uuid.UUID
		CreatedAt time.Time
	}
	var ids []row
	q := r.db.WithContext(ctx).Table("provider_balance_snapshots").
		Select("id, created_at").
		Where("created_at < ?", snapshot.Cutoff).
		Order("created_at ASC, id ASC").
		Limit(snapshot.BatchSize)
	if lastID != uuid.Nil {
		q = q.Where("(created_at, id) > (?, ?)", lastCreatedAt, lastID.String())
	}
	if err := q.Find(&ids).Error; err != nil {
		return 0, time.Time{}, uuid.Nil, false, fmt.Errorf("balance_snapshots cursor: %w", err)
	}
	if len(ids) == 0 {
		return 0, time.Time{}, uuid.Nil, false, nil
	}
	idList := make([]uuid.UUID, len(ids))
	for i, row := range ids {
		idList[i] = row.ID
	}
	res := r.db.WithContext(ctx).Table("provider_balance_snapshots").Where("id IN ?", idList).Delete(nil)
	if res.Error != nil {
		return 0, time.Time{}, uuid.Nil, false, fmt.Errorf("balance_snapshots delete: %w", res.Error)
	}
	last := ids[len(ids)-1]
	return res.RowsAffected, last.CreatedAt, last.ID, len(ids) == snapshot.BatchSize, nil
}
