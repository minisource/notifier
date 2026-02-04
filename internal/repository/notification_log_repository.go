package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
)

type NotificationLogRepository interface {
	Create(ctx context.Context, log *models.NotificationLog) error
	GetByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]*models.NotificationLog, error)
	GetByNotificationIDWithLimit(ctx context.Context, notificationID uuid.UUID, limit int) ([]*models.NotificationLog, error)
}

type notificationLogRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

func NewNotificationLogRepository(db *gorm.DB, logger logging.Logger) NotificationLogRepository {
	return &notificationLogRepository{
		db:     db,
		logger: logger,
	}
}

func (r *notificationLogRepository) Create(ctx context.Context, log *models.NotificationLog) error {
	r.logger.Debug(logging.Postgres, logging.Insert, "Creating notification log", map[logging.ExtraKey]interface{}{
		"notificationId": log.NotificationID,
		"action":         log.Action,
	})

	result := r.db.WithContext(ctx).Create(log)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Insert, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"notificationId": log.NotificationID,
		})
		return result.Error
	}

	return nil
}

func (r *notificationLogRepository) GetByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]*models.NotificationLog, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting logs by notification ID", map[logging.ExtraKey]interface{}{
		"notificationId": notificationID,
	})

	var logs []*models.NotificationLog
	result := r.db.WithContext(ctx).
		Where("notification_id = ?", notificationID).
		Order("created_at ASC").
		Find(&logs)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"notificationId": notificationID,
		})
		return nil, result.Error
	}

	return logs, nil
}

func (r *notificationLogRepository) GetByNotificationIDWithLimit(ctx context.Context, notificationID uuid.UUID, limit int) ([]*models.NotificationLog, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting logs by notification ID with limit", map[logging.ExtraKey]interface{}{
		"notificationId": notificationID,
		"limit":          limit,
	})

	var logs []*models.NotificationLog
	result := r.db.WithContext(ctx).
		Where("notification_id = ?", notificationID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"notificationId": notificationID,
		})
		return nil, result.Error
	}

	return logs, nil
}
