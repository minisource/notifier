package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Notification, int64, error)
	Update(ctx context.Context, notification *models.Notification) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.NotificationStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetPendingNotifications(ctx context.Context, limit int) ([]*models.Notification, error)
	GetRetryableNotifications(ctx context.Context, limit int) ([]*models.Notification, error)
	IncrementRetryCount(ctx context.Context, id uuid.UUID, nextRetryAt time.Time, errorMsg string) error
	MarkAsSent(ctx context.Context, id uuid.UUID, providerMsgID string) error
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	GetUnreadByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Notification, int64, error)
	GetByStatusAndType(ctx context.Context, status models.NotificationStatus, notifType models.NotificationType, limit int) ([]*models.Notification, error)

	// GetUnreadCountByUserID returns the count of unread notifications for a user
	GetUnreadCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// MarkAllAsReadByUserID marks all unread notifications as read for a user, returns count of updated rows
	MarkAllAsReadByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// MarkAsSeen marks a notification as seen (displayed to user)
	MarkAsSeen(ctx context.Context, id uuid.UUID) error

	// MarkAsClicked marks a notification as clicked (user interacted)
	MarkAsClicked(ctx context.Context, id uuid.UUID) error

	// MarkAsDelivered marks a notification as delivered (reached user's device)
	MarkAsDelivered(ctx context.Context, id uuid.UUID) error

	// GetInAppByUserID retrieves in-app notifications for a user (type='in_app')
	GetInAppByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Notification, int64, error)

	// GetByIDempotencyKey finds a notification by its idempotency key (prevents duplicates)
	GetByIDempotencyKey(ctx context.Context, key string) (*models.Notification, error)

	// MarkAsDeadLetter marks a notification as dead-letter (max retries exceeded)
	MarkAsDeadLetter(ctx context.Context, id uuid.UUID, errorMsg string) error

	// GetQueueDepth returns the count of pending + retrying notifications (queue depth metric)
	GetQueueDepth(ctx context.Context) (int64, error)

	// GetDigestedNotifications retrieves notifications with status=digested for a user
	GetDigestedNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]*models.Notification, error)

	// GetAllDigestedNotifications retrieves ALL notifications with status=digested (across all users)
	GetAllDigestedNotifications(ctx context.Context, limit int) ([]*models.Notification, error)

	// BulkUpdateStatus updates status for multiple notification IDs atomically
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status models.NotificationStatus) error

	// BulkRetry re-queues failed/dead notifications by ID for exactly one more
	// delivery attempt. It atomically sets status=pending and pins retry_count
	// to max_retries WITHOUT resetting it to zero, so a manual retry never
	// grants a fresh automatic-retry budget. Returns the number of rows
	// actually requeued.
	BulkRetry(ctx context.Context, ids []uuid.UUID) (int64, error)

	// RetryByID atomically re-queues a single failed/dead notification for one
	// more attempt (same semantics as BulkRetry). Returns true when a row was
	// actually updated; false when the notification was not in a retryable
	// state (or no longer exists) — no full-object Save, so it is safe against
	// concurrent worker updates.
	RetryByID(ctx context.Context, id uuid.UUID) (bool, error)

	// GetRetryableIDs returns the subset of the given IDs that are currently
	// in a manually retryable state (failed/dead). Used to log audit entries
	// only for notifications that actually qualify for retry.
	GetRetryableIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error)

	// GetAllRetryableIDs returns up to limit IDs of failed/dead notifications
	// (newest first) for the "retry all failed" operation.
	GetAllRetryableIDs(ctx context.Context, limit int) ([]uuid.UUID, error)

	// ListAll retrieves paginated notifications with optional filters (admin list)
	ListAll(ctx context.Context, filter NotificationListFilter) ([]*models.Notification, int64, error)

	// GetProviderStats aggregates delivery outcomes grouped by provider
	// (notifications.provider), used to enrich provider health/success-rate UI.
	GetProviderStats(ctx context.Context) ([]*ProviderStats, error)

	// SetProvider records which provider handled a notification (used after a
	// send attempt resolves so per-provider stats are accurate).
	SetProvider(ctx context.Context, id uuid.UUID, provider string) error

	// HoldForPause freezes a delivery into status=held WITHOUT consuming a
	// retry budget or touching retry_count. It records which pause generation
	// held it and why. No-op (false, nil) when the notification is already
	// terminal or already held.
	HoldForPause(ctx context.Context, id uuid.UUID, pauseVersion int64, reason string) (bool, error)
	// CountHeld returns the number of deliveries currently held by the pause.
	CountHeld(ctx context.Context) (int64, error)
	// CountActive returns the number of deliveries currently executing
	// (status sending/processing) — the active-attempt count for the pause UI.
	CountActive(ctx context.Context) (int64, error)
	// CountHeldRetries returns the number of held deliveries that were frozen
	// during a retry (retry_count > 0) — the retrying-held count for the UI.
	CountHeldRetries(ctx context.Context) (int64, error)
	// ReleaseHeld re-queues up to limit of the oldest held deliveries back to
	// pending so the poller/worker can resume them (controlled release, no
	// thundering herd). Returns the number released.
	ReleaseHeld(ctx context.Context, limit int) (int64, error)
	// ListHeld returns held deliveries oldest-first (for admin review).
	ListHeld(ctx context.Context, limit, offset int) ([]*models.Notification, int64, error)
	// RecoverStuckSending returns notifications stuck in sending/processing
	// longer than the cutoff back to pending so the DB poller re-enqueues them
	// after a worker crash (lease recovery). Never consumes retry budget and
	// never touches held/failed/terminal rows.
	RecoverStuckSending(ctx context.Context, cutoff time.Time) (int64, error)
}

// ProviderStats holds per-provider delivery statistics aggregated from the
// notifications table (grouped by the provider that handled each notification).
type ProviderStats struct {
	Provider         string     `gorm:"column:provider"`
	TotalSent        int64      `gorm:"column:total_sent"`
	TotalFailed      int64      `gorm:"column:total_failed"`
	AverageLatencyMs float64    `gorm:"column:avg_latency_ms"`
	LastSuccessAt    *time.Time `gorm:"column:last_success_at"`
	LastFailureAt    *time.Time `gorm:"column:last_failure_at"`
	LastError        string     `gorm:"column:last_error"`
	SuccessRate      float64    `gorm:"-"`
}

// NotificationListFilter represents filters for the admin notification list
type NotificationListFilter struct {
	Page          int
	PageSize      int
	Status        *models.NotificationStatus
	Type          *models.NotificationType
	UserID        *uuid.UUID
	TenantID      *uuid.UUID
	Search        string
	From          *time.Time
	To            *time.Time
	SortBy        string
	SortDirection string
}

type notificationRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

func NewNotificationRepository(db *gorm.DB, logger logging.Logger) NotificationRepository {
	return &notificationRepository{
		db:     db,
		logger: logger,
	}
}

func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	r.logger.Debug(logging.Postgres, logging.Insert, "Creating notification", map[logging.ExtraKey]interface{}{
		"userId": notification.UserID,
		"type":   notification.Type,
	})

	result := r.db.WithContext(ctx).Create(notification)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Insert, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"userId": notification.UserID,
		})
		return result.Error
	}

	r.logger.Debug(logging.Postgres, logging.Insert, "Notification created successfully", map[logging.ExtraKey]interface{}{
		"notificationId": notification.ID,
	})
	return nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting notification by ID", map[logging.ExtraKey]interface{}{
		"id": id,
	})

	var notification models.Notification
	result := r.db.WithContext(ctx).Preload("Template").First(&notification, "id = ?", id)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return nil, result.Error
	}

	return &notification, nil
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Notification, int64, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting notifications by user ID", map[logging.ExtraKey]interface{}{
		"userId": userID,
		"limit":  limit,
		"offset": offset,
	})

	var notifications []*models.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", userID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error(logging.Postgres, logging.Select, err.Error(), map[logging.ExtraKey]interface{}{
			"userId": userID,
		})
		return nil, 0, err
	}

	// Get paginated results
	result := query.Preload("Template").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"userId": userID,
		})
		return nil, 0, result.Error
	}

	r.logger.Debug(logging.Postgres, logging.Select, "Notifications retrieved successfully", map[logging.ExtraKey]interface{}{
		"userId": userID,
		"count":  len(notifications),
		"total":  total,
	})

	return notifications, total, nil
}

func (r *notificationRepository) Update(ctx context.Context, notification *models.Notification) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Updating notification", map[logging.ExtraKey]interface{}{
		"id": notification.ID,
	})

	result := r.db.WithContext(ctx).Save(notification)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": notification.ID,
		})
		return result.Error
	}

	return nil
}

func (r *notificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.NotificationStatus) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Updating notification status", map[logging.ExtraKey]interface{}{
		"id":     id,
		"status": status,
	})

	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}

	return nil
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug(logging.Postgres, logging.Delete, "Deleting notification", map[logging.ExtraKey]interface{}{
		"id": id,
	})

	result := r.db.WithContext(ctx).Delete(&models.Notification{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Delete, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}

	return nil
}

func (r *notificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]*models.Notification, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting pending notifications", map[logging.ExtraKey]interface{}{
		"limit": limit,
	})

	var notifications []*models.Notification
	result := r.db.WithContext(ctx).
		Where("status = ?", models.NotificationStatusPending).
		Where("(scheduled_at IS NULL OR scheduled_at <= ?)", time.Now()).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&notifications)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), nil)
		return nil, result.Error
	}

	r.logger.Debug(logging.Postgres, logging.Select, "Pending notifications retrieved", map[logging.ExtraKey]interface{}{
		"count": len(notifications),
	})

	return notifications, nil
}

func (r *notificationRepository) GetRetryableNotifications(ctx context.Context, limit int) ([]*models.Notification, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting retryable notifications", map[logging.ExtraKey]interface{}{
		"limit": limit,
	})

	var notifications []*models.Notification
	// Use `retry_count <= max_retries` (not `<`): the notification must be
	// re-enqueued one final time when retry_count == max_retries so the worker's
	// handleFailure can execute its terminal branch (MarkAsDeadLetter → dead).
	// With `<`, that final attempt was never picked up and notifications stayed
	// stuck in status=retrying forever after exhausting their retries.
	result := r.db.WithContext(ctx).
		Where("status = ?", models.NotificationStatusRetrying).
		Where("retry_count <= max_retries").
		Where("next_retry_at <= ?", time.Now()).
		Order("priority DESC, next_retry_at ASC").
		Limit(limit).
		Find(&notifications)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), nil)
		return nil, result.Error
	}

	r.logger.Debug(logging.Postgres, logging.Select, "Retryable notifications retrieved", map[logging.ExtraKey]interface{}{
		"count": len(notifications),
	})

	return notifications, nil
}

func (r *notificationRepository) IncrementRetryCount(ctx context.Context, id uuid.UUID, nextRetryAt time.Time, errorMsg string) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Incrementing retry count", map[logging.ExtraKey]interface{}{
		"id":          id,
		"nextRetryAt": nextRetryAt,
	})

	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"retry_count":   gorm.Expr("retry_count + 1"),
			"next_retry_at": nextRetryAt,
			"error_message": errorMsg,
			"status":        models.NotificationStatusRetrying,
		})

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}

	return nil
}

func (r *notificationRepository) MarkAsSent(ctx context.Context, id uuid.UUID, providerMsgID string) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Marking notification as sent", map[logging.ExtraKey]interface{}{
		"id":            id,
		"providerMsgId": providerMsgID,
	})

	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          models.NotificationStatusSent,
			"sent_at":         now,
			"provider_msg_id": providerMsgID,
		})

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}

	return nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Marking notification as read", map[logging.ExtraKey]interface{}{
		"id": id,
	})

	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Update("read_at", now)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}

	return nil
}

func (r *notificationRepository) GetUnreadByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Notification, int64, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting unread notifications", map[logging.ExtraKey]interface{}{
		"userId": userID,
		"limit":  limit,
		"offset": offset,
	})

	var notifications []*models.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error(logging.Postgres, logging.Select, err.Error(), map[logging.ExtraKey]interface{}{
			"userId": userID,
		})
		return nil, 0, err
	}

	// Get paginated results
	result := query.Preload("Template").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"userId": userID,
		})
		return nil, 0, result.Error
	}

	return notifications, total, nil
}

func (r *notificationRepository) GetUnreadCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting unread notification count", map[logging.ExtraKey]interface{}{
		"userId": userID,
	})

	var count int64
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"userId": userID,
		})
		return 0, result.Error
	}

	return count, nil
}

func (r *notificationRepository) GetByIDempotencyKey(ctx context.Context, key string) (*models.Notification, error) {
	if key == "" {
		return nil, nil
	}
	var notification models.Notification
	result := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&notification)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &notification, nil
}

func (r *notificationRepository) MarkAsDeadLetter(ctx context.Context, id uuid.UUID, errorMsg string) error {
	r.logger.Warn(logging.Postgres, logging.Update, "Marking notification as dead-letter (canceled)", map[logging.ExtraKey]interface{}{
		"id": id,
	})
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        models.NotificationStatusDead,
			"error_message": gorm.Expr("COALESCE(error_message, '') || '\n' || ?", "DEAD_LETTER: "+errorMsg),
		})
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}
	return nil
}

func (r *notificationRepository) GetQueueDepth(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("status IN ?", []models.NotificationStatus{models.NotificationStatusPending, models.NotificationStatusRetrying}).
		Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func (r *notificationRepository) MarkAsSeen(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND seen_at IS NULL", id).
		Update("seen_at", now).Error
}

func (r *notificationRepository) MarkAsClicked(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND clicked_at IS NULL", id).
		Update("clicked_at", now).Error
}

func (r *notificationRepository) MarkAsDelivered(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Update("delivered_at", now).Error
}

func (r *notificationRepository) GetInAppByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Notification, int64, error) {
	var notifications []*models.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND type = ?", userID, models.NotificationTypeInApp)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifications)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	return notifications, total, nil
}

func (r *notificationRepository) MarkAllAsReadByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	r.logger.Debug(logging.Postgres, logging.Update, "Marking all notifications as read", map[logging.ExtraKey]interface{}{
		"userId": userID,
	})

	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", now)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"userId": userID,
		})
		return 0, result.Error
	}

	r.logger.Info(logging.Postgres, logging.Update, "Marked all notifications as read", map[logging.ExtraKey]interface{}{
		"userId": userID,
		"count":  result.RowsAffected,
	})

	return result.RowsAffected, nil
}

func (r *notificationRepository) GetDigestedNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, models.NotificationStatusDigested).
		Order("created_at ASC").
		Limit(limit).
		Find(&notifications)
	if result.Error != nil {
		return nil, result.Error
	}
	return notifications, nil
}

func (r *notificationRepository) GetAllDigestedNotifications(ctx context.Context, limit int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	result := r.db.WithContext(ctx).
		Where("status = ?", models.NotificationStatusDigested).
		Order("user_id ASC, created_at ASC").
		Limit(limit).
		Find(&notifications)
	if result.Error != nil {
		return nil, result.Error
	}
	return notifications, nil
}

func (r *notificationRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status models.NotificationStatus) error {
	if len(ids) == 0 {
		return nil
	}
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id IN ?", ids).
		Update("status", status)
	return result.Error
}

// retryUpdateMap is shared by BulkRetry and RetryAllFailed so both paths
// requeue with identical semantics: status=pending (so the worker/DB poller
// picks the notification up immediately), retry_count = max_retries (a manual
// retry grants exactly ONE attempt — never a fresh automatic-retry budget), and
// next_retry_at cleared (no backoff gate on the manual retry).
func retryUpdateMap(now time.Time) map[string]interface{} {
	return map[string]interface{}{
		"status":        models.NotificationStatusPending,
		"retry_count":   gorm.Expr("max_retries"),
		"next_retry_at": nil,
		"updated_at":    now,
	}
}

func (r *notificationRepository) BulkRetry(ctx context.Context, ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id IN ?", ids).
		Where("status IN ?", []models.NotificationStatus{models.NotificationStatusFailed, models.NotificationStatusDead}).
		Updates(retryUpdateMap(time.Now()))
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, "Failed to bulk retry notifications", map[logging.ExtraKey]interface{}{
			"ids":   len(ids),
			"error": result.Error.Error(),
		})
		return 0, result.Error
	}
	r.logger.Info(logging.Postgres, logging.Update, "Bulk retried notifications", map[logging.ExtraKey]interface{}{
		"requested": len(ids),
		"retried":   result.RowsAffected,
	})
	return result.RowsAffected, nil
}

func (r *notificationRepository) RetryByID(ctx context.Context, id uuid.UUID) (bool, error) {
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Where("status IN ?", []models.NotificationStatus{models.NotificationStatusFailed, models.NotificationStatusDead}).
		Updates(retryUpdateMap(time.Now()))
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, "Failed to retry notification by id", map[logging.ExtraKey]interface{}{
			"id":    id,
			"error": result.Error.Error(),
		})
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *notificationRepository) GetRetryableIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, nil
	}
	var retryable []uuid.UUID
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id IN ?", ids).
		Where("status IN ?", []models.NotificationStatus{models.NotificationStatusFailed, models.NotificationStatusDead}).
		Pluck("id", &retryable)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, "Failed to find retryable notification ids", map[logging.ExtraKey]interface{}{
			"error": result.Error.Error(),
		})
		return nil, result.Error
	}
	return retryable, nil
}

func (r *notificationRepository) GetAllRetryableIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("status IN ?", []models.NotificationStatus{models.NotificationStatusFailed, models.NotificationStatusDead}).
		Order("created_at DESC").
		Limit(limit).
		Pluck("id", &ids)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, "Failed to list retryable notification ids", map[logging.ExtraKey]interface{}{
			"error": result.Error.Error(),
		})
		return nil, result.Error
	}
	return ids, nil
}

func (r *notificationRepository) ListAll(ctx context.Context, filter NotificationListFilter) ([]*models.Notification, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Notification{})

	// Apply filters
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Type != nil && *filter.Type != "" {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.TenantID != nil {
		query = query.Where("tenant_id = ?", *filter.TenantID)
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("created_at <= ?", *filter.To)
	}

	// Search: match against subject, body, recipientEmail, recipientPhone, userId
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where(
			"subject ILIKE ? OR body ILIKE ? OR recipient_email ILIKE ? OR recipient_phone ILIKE ? OR CAST(user_id AS TEXT) ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Count total matching records before pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sort
	allowedSortFields := map[string]bool{
		"created_at": true, "updated_at": true, "sent_at": true,
		"priority": true, "status": true,
	}
	sortBy := "created_at"
	if filter.SortBy != "" && allowedSortFields[filter.SortBy] {
		sortBy = filter.SortBy
	}
	sortDir := "DESC"
	if filter.SortDirection == "asc" {
		sortDir = "ASC"
	}

	// Paginate
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var notifications []*models.Notification
	result := query.
		Preload("Template").
		Order(sortBy + " " + sortDir).
		Limit(pageSize).
		Offset(offset).
		Find(&notifications)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return notifications, total, nil
}

func (r *notificationRepository) SetProvider(ctx context.Context, id uuid.UUID, provider string) error {
	if provider == "" {
		return nil
	}
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Update("provider", provider)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, "Failed to set notification provider", map[logging.ExtraKey]interface{}{
			"id":       id,
			"provider": provider,
			"error":    result.Error.Error(),
		})
		return result.Error
	}
	return nil
}

func (r *notificationRepository) GetProviderStats(ctx context.Context) ([]*ProviderStats, error) {
	const q = `
SELECT n.provider AS provider,
       COUNT(*) FILTER (WHERE n.status IN ('sent','delivered')) AS total_sent,
       COUNT(*) FILTER (WHERE n.status IN ('failed','dead')) AS total_failed,
       COALESCE(AVG(EXTRACT(EPOCH FROM (n.sent_at - n.created_at)) * 1000), 0) AS avg_latency_ms,
       MAX(n.sent_at) AS last_success_at,
       MAX(n.failed_at) AS last_failure_at,
       (SELECT n2.error_message FROM notifications n2
         WHERE n2.provider = n.provider AND n2.status IN ('failed','dead')
         ORDER BY n2.created_at DESC LIMIT 1) AS last_error
FROM notifications n
WHERE n.provider IS NOT NULL AND n.provider != ''
GROUP BY n.provider`

	var rows []*ProviderStats
	if err := r.db.WithContext(ctx).Raw(q).Scan(&rows).Error; err != nil {
		r.logger.Error(logging.Postgres, logging.Select, "Failed to aggregate provider stats", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	for _, s := range rows {
		total := s.TotalSent + s.TotalFailed
		if total > 0 {
			s.SuccessRate = float64(s.TotalSent) / float64(total) * 100
		}
	}
	return rows, nil
}

func (r *notificationRepository) HoldForPause(ctx context.Context, id uuid.UUID, pauseVersion int64, reason string) (bool, error) {
	if pauseVersion <= 0 {
		pauseVersion = 1
	}
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Where("status IN ?", []models.NotificationStatus{
			models.NotificationStatusPending,
			models.NotificationStatusQueued,
			models.NotificationStatusRetrying,
			models.NotificationStatusProcessing,
			models.NotificationStatusSending,
		}).
		Updates(map[string]interface{}{
			"status":         models.NotificationStatusHeld,
			"pause_version":  pauseVersion,
			"held_reason":    reason,
			"held_at":        now,
			"updated_at":     now,
		})
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, "Failed to hold notification for pause", map[logging.ExtraKey]interface{}{
			"id": id, "error": result.Error.Error(),
		})
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *notificationRepository) CountHeld(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("status = ?", models.NotificationStatusHeld).Count(&count).Error
	return count, err
}

func (r *notificationRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("status IN ?", []models.NotificationStatus{
			models.NotificationStatusSending,
			models.NotificationStatusProcessing,
		}).Count(&count).Error
	return count, err
}

func (r *notificationRepository) CountHeldRetries(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("status = ?", models.NotificationStatusHeld).
		Where("retry_count > 0").Count(&count).Error
	return count, err
}

func (r *notificationRepository) ReleaseHeld(ctx context.Context, limit int) (int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	// Re-queue the OLDEST held deliveries first; they were blocked the longest.
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id IN (?)",
			r.db.Model(&models.Notification{}).
				Select("id").
				Where("status = ?", models.NotificationStatusHeld).
				Order("held_at ASC, created_at ASC").
				Limit(limit),
		).
		Updates(map[string]interface{}{
			"status":        models.NotificationStatusPending,
			"updated_at":    time.Now(),
		})
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, "Failed to release held notifications", map[logging.ExtraKey]interface{}{
			"error": result.Error.Error(),
		})
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (r *notificationRepository) ListHeld(ctx context.Context, limit, offset int) ([]*models.Notification, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var total int64
	q := r.db.WithContext(ctx).Model(&models.Notification{}).Where("status = ?", models.NotificationStatusHeld)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []*models.Notification
	err := r.db.WithContext(ctx).Where("status = ?", models.NotificationStatusHeld).
		Order("held_at ASC, created_at ASC").Limit(limit).Offset(offset).Find(&out).Error
	return out, total, err
}

// RecoverStuckSending resets deliveries stuck in sending/processing past the
// cutoff back to pending (worker-crash lease recovery). Only rows whose
// updated_at is older than the cutoff are touched, so a genuinely slow
// provider call is never interrupted; the poller picks them up afterwards.
func (r *notificationRepository) RecoverStuckSending(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("status IN ?", []models.NotificationStatus{
			models.NotificationStatusSending,
			models.NotificationStatusProcessing,
		}).
		Where("updated_at < ?", cutoff).
		Updates(map[string]interface{}{
			"status":     models.NotificationStatusPending,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		r.logger.Warn(logging.Postgres, logging.Update, "Failed to recover stuck sending notifications", map[logging.ExtraKey]interface{}{
			"error": result.Error.Error(),
		})
		return 0, result.Error
	}
	if result.RowsAffected > 0 {
		r.logger.Info(logging.Postgres, logging.Update, "Recovered stuck sending notifications (lease recovery)", map[logging.ExtraKey]interface{}{
			"count": result.RowsAffected,
		})
	}
	return result.RowsAffected, nil
}

func (r *notificationRepository) GetByStatusAndType(ctx context.Context, status models.NotificationStatus, notifType models.NotificationType, limit int) ([]*models.Notification, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting notifications by status and type", map[logging.ExtraKey]interface{}{
		"status": status,
		"type":   notifType,
		"limit":  limit,
	})

	var notifications []*models.Notification
	result := r.db.WithContext(ctx).
		Where("status = ? AND type = ?", status, notifType).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"status": status,
			"type":   notifType,
		})
		return nil, result.Error
	}

	return notifications, nil
}
