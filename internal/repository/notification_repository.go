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
	result := r.db.WithContext(ctx).
		Where("status = ?", models.NotificationStatusRetrying).
		Where("retry_count < max_retries").
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
