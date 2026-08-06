package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
)

// ProviderAttemptRepository persists durable provider attempt records and
// their lifecycle events. It satisfies attemptlog.AttemptRepository.
type ProviderAttemptRepository interface {
	CreateAttempt(ctx context.Context, attempt *models.ProviderAttempt) error
	UpdateAttempt(ctx context.Context, attempt *models.ProviderAttempt) error
	GetAttemptByID(ctx context.Context, id uuid.UUID) (*models.ProviderAttempt, error)
	AddEvent(ctx context.Context, event *models.ProviderAttemptEvent) error
	ListEventsByAttempt(ctx context.Context, attemptID uuid.UUID) ([]*models.ProviderAttemptEvent, error)
	ListByNotification(ctx context.Context, notificationID uuid.UUID) ([]*models.ProviderAttempt, error)
	// List returns paginated attempts newest-first matching the filter.
	List(ctx context.Context, filter ProviderAttemptFilter) ([]*models.ProviderAttempt, int64, error)
	// PurgeBodies clears captured request/response bodies for attempts older
	// than bodyRetention (metadata preserved).
	PurgeBodies(ctx context.Context, olderThan time.Time) (int64, error)
	// DeleteExpired hard-deletes attempts (and their events) older than
	// metadataRetention.
	DeleteExpired(ctx context.Context, olderThan time.Time) (int64, error)
}

// ProviderAttemptFilter carries list query options.
type ProviderAttemptFilter struct {
	Page               int
	PageSize           int
	NotificationID     *uuid.UUID
	ProviderAccountID  *uuid.UUID
	TenantID           *uuid.UUID
	Channel            string
	Provider           string
	Status             string
	ProviderMessageID  string
	RequestID          string
	CorrelationID      string
	From               *time.Time
	To                 *time.Time
}

type providerAttemptRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

// NewProviderAttemptRepository creates the repository.
func NewProviderAttemptRepository(db *gorm.DB, logger logging.Logger) ProviderAttemptRepository {
	return &providerAttemptRepository{db: db, logger: logger}
}

func (r *providerAttemptRepository) CreateAttempt(ctx context.Context, attempt *models.ProviderAttempt) error {
	return r.db.WithContext(ctx).Create(attempt).Error
}

func (r *providerAttemptRepository) UpdateAttempt(ctx context.Context, attempt *models.ProviderAttempt) error {
	return r.db.WithContext(ctx).Model(attempt).Updates(map[string]interface{}{
		"status":                    attempt.Status,
		"provider_status":           attempt.ProviderStatus,
		"provider_message_id":       attempt.ProviderMessageID,
		"provider_error_code":       attempt.ProviderErrorCode,
		"normalized_error_kind":     attempt.NormalizedErrorKind,
		"normalized_error_code":     attempt.NormalizedErrorCode,
		"normalized_error_message":  attempt.NormalizedErrorMessage,
		"retryable":                 attempt.Retryable,
		"request_method":            attempt.RequestMethod,
		"request_url_sanitized":     attempt.RequestURLSanitized,
		"request_headers_sanitized": attempt.RequestHeadersSanitized,
		"request_body_sanitized":    attempt.RequestBodySanitized,
		"request_size_bytes":        attempt.RequestSizeBytes,
		"response_status_code":      attempt.ResponseStatusCode,
		"response_headers_sanitized": attempt.ResponseHeadersSanitized,
		"response_body_sanitized":   attempt.ResponseBodySanitized,
		"response_size_bytes":       attempt.ResponseSizeBytes,
		"body_truncated":            attempt.BodyTruncated,
		"original_size_bytes":       attempt.OriginalSizeBytes,
		"captured_size_bytes":       attempt.CapturedSizeBytes,
		"started_at":                attempt.StartedAt,
		"completed_at":              attempt.CompletedAt,
		"duration_ms":               attempt.DurationMs,
		"updated_at":                time.Now().UTC(),
	}).Error
}

func (r *providerAttemptRepository) GetAttemptByID(ctx context.Context, id uuid.UUID) (*models.ProviderAttempt, error) {
	var attempt models.ProviderAttempt
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&attempt).Error
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *providerAttemptRepository) AddEvent(ctx context.Context, event *models.ProviderAttemptEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *providerAttemptRepository) ListEventsByAttempt(ctx context.Context, attemptID uuid.UUID) ([]*models.ProviderAttemptEvent, error) {
	var events []*models.ProviderAttemptEvent
	err := r.db.WithContext(ctx).
		Where("attempt_id = ?", attemptID).
		Order("occurred_at ASC").
		Find(&events).Error
	return events, err
}

func (r *providerAttemptRepository) ListByNotification(ctx context.Context, notificationID uuid.UUID) ([]*models.ProviderAttempt, error) {
	var attempts []*models.ProviderAttempt
	err := r.db.WithContext(ctx).
		Where("notification_id = ?", notificationID).
		Order("attempt_number ASC, fallback_sequence ASC, created_at ASC").
		Find(&attempts).Error
	return attempts, err
}

func (r *providerAttemptRepository) List(ctx context.Context, filter ProviderAttemptFilter) ([]*models.ProviderAttempt, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	q := r.db.WithContext(ctx).Model(&models.ProviderAttempt{})
	if filter.NotificationID != nil {
		q = q.Where("notification_id = ?", *filter.NotificationID)
	}
	if filter.ProviderAccountID != nil {
		q = q.Where("provider_account_id = ?", *filter.ProviderAccountID)
	}
	if filter.TenantID != nil {
		q = q.Where("tenant_id = ?", *filter.TenantID)
	}
	if filter.Channel != "" {
		q = q.Where("channel = ?", filter.Channel)
	}
	if filter.Provider != "" {
		q = q.Where("provider = ?", filter.Provider)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.ProviderMessageID != "" {
		q = q.Where("provider_message_id = ?", filter.ProviderMessageID)
	}
	if filter.RequestID != "" {
		q = q.Where("request_id = ?", filter.RequestID)
	}
	if filter.CorrelationID != "" {
		q = q.Where("correlation_id = ?", filter.CorrelationID)
	}
	if filter.From != nil {
		q = q.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("created_at <= ?", *filter.To)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var attempts []*models.ProviderAttempt
	err := q.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&attempts).Error
	if err != nil {
		return nil, 0, err
	}
	return attempts, total, nil
}

func (r *providerAttemptRepository) PurgeBodies(ctx context.Context, olderThan time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Model(&models.ProviderAttempt{}).
		Where("created_at < ? AND (request_body_sanitized <> '' OR response_body_sanitized <> '')", olderThan).
		Updates(map[string]interface{}{
			"request_body_sanitized":   "",
			"response_body_sanitized":  "",
			"request_headers_sanitized": "{}",
			"response_headers_sanitized": "{}",
			"request_size_bytes":       0,
			"response_size_bytes":      0,
			"body_truncated":           false,
			"original_size_bytes":      0,
			"captured_size_bytes":      0,
			"updated_at":               time.Now().UTC(),
		})
	return res.RowsAffected, res.Error
}

func (r *providerAttemptRepository) DeleteExpired(ctx context.Context, olderThan time.Time) (int64, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&models.ProviderAttempt{}).
		Where("created_at < ?", olderThan).
		Pluck("id", &ids).Error
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	// Delete events first (FK safety / cleanliness), then hard-delete attempts.
	// Unscoped() is required: ProviderAttempt embeds gorm.DeletedAt so a plain
	// Delete would be a soft delete and never reclaim storage — defeating the
	// purpose of a retention cleanup job.
	evErr := r.db.WithContext(ctx).Where("attempt_id IN ?", ids).Delete(&models.ProviderAttemptEvent{}).Error
	if evErr != nil {
		return 0, evErr
	}
	res := r.db.WithContext(ctx).Unscoped().Where("id IN ?", ids).Delete(&models.ProviderAttempt{}).Error
	if res != nil {
		return 0, res
	}
	return int64(len(ids)), nil
}
