package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
)

type NotificationPreferenceRepository interface {
	Create(ctx context.Context, pref *models.NotificationPreference) error
	GetByUserIDAndType(ctx context.Context, userID uuid.UUID, notifType models.NotificationType) (*models.NotificationPreference, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.NotificationPreference, error)
	Update(ctx context.Context, pref *models.NotificationPreference) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
	Upsert(ctx context.Context, pref *models.NotificationPreference) error
}

type notificationPreferenceRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

func NewNotificationPreferenceRepository(db *gorm.DB, logger logging.Logger) NotificationPreferenceRepository {
	return &notificationPreferenceRepository{
		db:     db,
		logger: logger,
	}
}

func (r *notificationPreferenceRepository) Create(ctx context.Context, pref *models.NotificationPreference) error {
	r.logger.Debug(logging.Postgres, logging.Insert, "Creating notification preference", map[logging.ExtraKey]interface{}{
		"userId": pref.UserID,
		"type":   pref.Type,
	})

	result := r.db.WithContext(ctx).Create(pref)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Insert, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"userId": pref.UserID,
		})
		return result.Error
	}

	return nil
}

func (r *notificationPreferenceRepository) GetByUserIDAndType(ctx context.Context, userID uuid.UUID, notifType models.NotificationType) (*models.NotificationPreference, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting preference by user ID and type", map[logging.ExtraKey]interface{}{
		"userId": userID,
		"type":   notifType,
	})

	var pref models.NotificationPreference
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, notifType).
		First(&pref)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			r.logger.Debug(logging.Postgres, logging.Select, "No preference found, will use defaults", map[logging.ExtraKey]interface{}{
				"userId": userID,
				"type":   notifType,
			})
		} else {
			r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
				"userId": userID,
				"type":   notifType,
			})
		}
		return nil, result.Error
	}

	return &pref, nil
}

func (r *notificationPreferenceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.NotificationPreference, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting all preferences for user", map[logging.ExtraKey]interface{}{
		"userId": userID,
	})

	var prefs []*models.NotificationPreference
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&prefs)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"userId": userID,
		})
		return nil, result.Error
	}

	return prefs, nil
}

func (r *notificationPreferenceRepository) Update(ctx context.Context, pref *models.NotificationPreference) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Updating notification preference", map[logging.ExtraKey]interface{}{
		"id": pref.ID,
	})

	result := r.db.WithContext(ctx).Save(pref)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": pref.ID,
		})
		return result.Error
	}

	return nil
}

func (r *notificationPreferenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug(logging.Postgres, logging.Delete, "Deleting notification preference", map[logging.ExtraKey]interface{}{
		"id": id,
	})

	result := r.db.WithContext(ctx).Delete(&models.NotificationPreference{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Delete, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}

	return nil
}

func (r *notificationPreferenceRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	r.logger.Debug(logging.Postgres, logging.Delete, "Deleting all preferences for user", map[logging.ExtraKey]interface{}{
		"userId": userID,
	})

	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.NotificationPreference{})
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Delete, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"userId": userID,
		})
		return result.Error
	}

	r.logger.Debug(logging.Postgres, logging.Delete, "Deleted preferences successfully", map[logging.ExtraKey]interface{}{
		"userId": userID,
		"count":  result.RowsAffected,
	})

	return nil
}

func (r *notificationPreferenceRepository) Upsert(ctx context.Context, pref *models.NotificationPreference) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Upserting notification preference", map[logging.ExtraKey]interface{}{
		"userId": pref.UserID,
		"type":   pref.Type,
	})

	// Try to find existing preference
	var existing models.NotificationPreference
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", pref.UserID, pref.Type).
		First(&existing)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new preference
		return r.Create(ctx, pref)
	} else if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"userId": pref.UserID,
		})
		return result.Error
	}

	// Update existing preference
	pref.ID = existing.ID
	return r.Update(ctx, pref)
}
