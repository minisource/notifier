package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProviderBalanceRepository persists provider account balance snapshots, the
// current health state, and deduplicated credit alerts.
type ProviderBalanceRepository interface {
	// CreateSnapshot appends one point-in-time observation.
	CreateSnapshot(ctx context.Context, s *models.ProviderBalanceSnapshot) error
	// ListSnapshots returns snapshots newest-first for a provider, optionally
	// bounded by time.
	ListSnapshots(ctx context.Context, providerID uuid.UUID, from, to *time.Time, limit int) ([]*models.ProviderBalanceSnapshot, error)
	// GetLatestSnapshot returns the most recent snapshot for a provider.
	GetLatestSnapshot(ctx context.Context, providerID uuid.UUID) (*models.ProviderBalanceSnapshot, error)

	// UpsertHealth upserts the current health row (keyed by provider_id).
	UpsertHealth(ctx context.Context, h *models.ProviderAccountHealth) error
	// GetHealth returns the health row for a provider (nil when absent).
	GetHealth(ctx context.Context, providerID uuid.UUID) (*models.ProviderAccountHealth, error)
	// ListHealth returns health rows for all providers (optionally tenant-scoped).
	ListHealth(ctx context.Context, tenantID *uuid.UUID) ([]*models.ProviderAccountHealth, error)

	// CreateAlert persists an alert occurrence.
	CreateAlert(ctx context.Context, a *models.ProviderCreditAlert) error
	// UpdateAlert updates an alert occurrence.
	UpdateAlert(ctx context.Context, a *models.ProviderCreditAlert) error
	// GetActiveAlert returns the currently-active alert of the given type for a
	// provider (nil when none) — the dedup guard.
	GetActiveAlert(ctx context.Context, providerID uuid.UUID, alertType string) (*models.ProviderCreditAlert, error)
	// ListAlerts returns alerts for a provider, optionally filtered by status,
	// newest-first.
	ListAlerts(ctx context.Context, providerID uuid.UUID, status string, limit int) ([]*models.ProviderCreditAlert, error)
	// ListAllAlerts returns alerts across providers (admin), newest-first.
	ListAllAlerts(ctx context.Context, status string, tenantID *uuid.UUID, limit int) ([]*models.ProviderCreditAlert, error)
	// ResolveAlertsForType resolves all active alerts of a type for a provider
	// (used on recovery). Returns the number resolved.
	ResolveAlertsForType(ctx context.Context, providerID uuid.UUID, alertType, reason string) (int64, error)

	// DeleteExpiredSnapshots purges snapshots older than retention (metadata +
	// values; health row is untouched).
	DeleteExpiredSnapshots(ctx context.Context, olderThan time.Time) (int64, error)
}

type providerBalanceRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

// NewProviderBalanceRepository creates the repository.
func NewProviderBalanceRepository(db *gorm.DB, logger logging.Logger) ProviderBalanceRepository {
	return &providerBalanceRepository{db: db, logger: logger}
}

func (r *providerBalanceRepository) CreateSnapshot(ctx context.Context, s *models.ProviderBalanceSnapshot) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *providerBalanceRepository) ListSnapshots(ctx context.Context, providerID uuid.UUID, from, to *time.Time, limit int) ([]*models.ProviderBalanceSnapshot, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.db.WithContext(ctx).Where("provider_id = ?", providerID)
	if from != nil {
		q = q.Where("fetched_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("fetched_at <= ?", *to)
	}
	var out []*models.ProviderBalanceSnapshot
	err := q.Order("fetched_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

func (r *providerBalanceRepository) GetLatestSnapshot(ctx context.Context, providerID uuid.UUID) (*models.ProviderBalanceSnapshot, error) {
	var s models.ProviderBalanceSnapshot
	err := r.db.WithContext(ctx).Where("provider_id = ?", providerID).Order("fetched_at DESC").First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *providerBalanceRepository) UpsertHealth(ctx context.Context, h *models.ProviderAccountHealth) error {
	// UpdateAll would also overwrite created_at on every refresh; assign the
	// mutable columns explicitly to preserve the original creation time.
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "provider_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"tenant_id", "provider", "channel", "capability_mode",
			"health_level", "latest_alert_level", "alert_updated_at",
			"balance_value", "balance_unit", "currency", "quota_remaining",
			"quota_limit", "usage_percent", "is_estimated", "is_manual", "source",
			"latest_snapshot_id", "last_successful_refresh_at", "last_refresh_attempt_at",
			"next_scheduled_refresh_at", "consecutive_failures", "last_error_kind",
			"last_error_message", "refresh_lock_until", "updated_at",
		}),
	}).Create(h).Error
}

func (r *providerBalanceRepository) GetHealth(ctx context.Context, providerID uuid.UUID) (*models.ProviderAccountHealth, error) {
	var h models.ProviderAccountHealth
	err := r.db.WithContext(ctx).Where("provider_id = ?", providerID).First(&h).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *providerBalanceRepository) ListHealth(ctx context.Context, tenantID *uuid.UUID) ([]*models.ProviderAccountHealth, error) {
	q := r.db.WithContext(ctx)
	if tenantID != nil {
		q = q.Where("tenant_id IS NULL OR tenant_id = ?", *tenantID)
	}
	var out []*models.ProviderAccountHealth
	err := q.Order("provider ASC, provider_id ASC").Find(&out).Error
	return out, err
}

func (r *providerBalanceRepository) CreateAlert(ctx context.Context, a *models.ProviderCreditAlert) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *providerBalanceRepository) UpdateAlert(ctx context.Context, a *models.ProviderCreditAlert) error {
	return r.db.WithContext(ctx).Model(a).Updates(map[string]interface{}{
		"status":           a.Status,
		"last_triggered_at": a.LastTriggeredAt,
		"repeat_count":     a.RepeatCount,
		"acknowledged_at":  a.AcknowledgedAt,
		"acknowledged_by":  a.AcknowledgedBy,
		"resolved_at":      a.ResolvedAt,
		"resolved_reason":  a.ResolvedReason,
		"updated_at":       time.Now().UTC(),
	}).Error
}

func (r *providerBalanceRepository) GetActiveAlert(ctx context.Context, providerID uuid.UUID, alertType string) (*models.ProviderCreditAlert, error) {
	var a models.ProviderCreditAlert
	err := r.db.WithContext(ctx).
		Where("provider_id = ? AND alert_type = ? AND status = ?", providerID, alertType, models.AlertStatusActive).
		Order("created_at DESC").First(&a).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *providerBalanceRepository) ListAlerts(ctx context.Context, providerID uuid.UUID, status string, limit int) ([]*models.ProviderCreditAlert, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.db.WithContext(ctx).Where("provider_id = ?", providerID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var out []*models.ProviderCreditAlert
	err := q.Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

func (r *providerBalanceRepository) ListAllAlerts(ctx context.Context, status string, tenantID *uuid.UUID, limit int) ([]*models.ProviderCreditAlert, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.db.WithContext(ctx)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if tenantID != nil {
		q = q.Where("tenant_id IS NULL OR tenant_id = ?", *tenantID)
	}
	var out []*models.ProviderCreditAlert
	err := q.Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

func (r *providerBalanceRepository) ResolveAlertsForType(ctx context.Context, providerID uuid.UUID, alertType, reason string) (int64, error) {
	res := r.db.WithContext(ctx).Model(&models.ProviderCreditAlert{}).
		Where("provider_id = ? AND alert_type = ? AND status = ?", providerID, alertType, models.AlertStatusActive).
		Updates(map[string]interface{}{
			"status":          models.AlertStatusResolved,
			"resolved_at":     time.Now().UTC(),
			"resolved_reason": reason,
			"updated_at":      time.Now().UTC(),
		})
	return res.RowsAffected, res.Error
}

func (r *providerBalanceRepository) DeleteExpiredSnapshots(ctx context.Context, olderThan time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Where("created_at < ?", olderThan).Delete(&models.ProviderBalanceSnapshot{})
	return res.RowsAffected, res.Error
}
