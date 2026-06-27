package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
)

type NotificationTemplateRepository interface {
	Create(ctx context.Context, template *models.NotificationTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.NotificationTemplate, error)
	GetByName(ctx context.Context, name string, notifType models.NotificationType) (*models.NotificationTemplate, error)
	GetByKey(ctx context.Context, key string) (*models.NotificationTemplate, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.NotificationTemplate, int64, error)
	GetByType(ctx context.Context, notifType models.NotificationType) ([]*models.NotificationTemplate, error)
	Update(ctx context.Context, template *models.NotificationTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetActive(ctx context.Context) ([]*models.NotificationTemplate, error)

	// GetByKeyAndLocale retrieves an active template by key and locale (with fallback to 'en')
	GetByKeyAndLocale(ctx context.Context, key string, locale string) (*models.NotificationTemplate, error)
}

type notificationTemplateRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

func NewNotificationTemplateRepository(db *gorm.DB, logger logging.Logger) NotificationTemplateRepository {
	return &notificationTemplateRepository{
		db:     db,
		logger: logger,
	}
}

func (r *notificationTemplateRepository) Create(ctx context.Context, template *models.NotificationTemplate) error {
	r.logger.Debug(logging.Postgres, logging.Insert, "Creating notification template", map[logging.ExtraKey]interface{}{
		"name": template.Name,
		"type": template.Type,
	})

	result := r.db.WithContext(ctx).Create(template)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Insert, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"name": template.Name,
		})
		return result.Error
	}

	r.logger.Debug(logging.Postgres, logging.Insert, "Template created successfully", map[logging.ExtraKey]interface{}{
		"templateId": template.ID,
	})
	return nil
}

func (r *notificationTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.NotificationTemplate, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting template by ID", map[logging.ExtraKey]interface{}{
		"id": id,
	})

	var template models.NotificationTemplate
	result := r.db.WithContext(ctx).First(&template, "id = ?", id)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return nil, result.Error
	}

	return &template, nil
}

func (r *notificationTemplateRepository) GetByName(ctx context.Context, name string, notifType models.NotificationType) (*models.NotificationTemplate, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting template by name and type", map[logging.ExtraKey]interface{}{
		"name": name,
		"type": notifType,
	})

	var template models.NotificationTemplate
	result := r.db.WithContext(ctx).
		Where("name = ? AND type = ?", name, notifType).
		First(&template)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"name": name,
			"type": notifType,
		})
		return nil, result.Error
	}

	return &template, nil
}

func (r *notificationTemplateRepository) GetByKey(ctx context.Context, key string) (*models.NotificationTemplate, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting template by key", map[logging.ExtraKey]interface{}{
		"key": key,
	})

	var template models.NotificationTemplate
	result := r.db.WithContext(ctx).
		Where("key = ?", key).
		First(&template)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"key": key,
		})
		return nil, result.Error
	}

	return &template, nil
}

func (r *notificationTemplateRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.NotificationTemplate, int64, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting all templates", map[logging.ExtraKey]interface{}{
		"limit":  limit,
		"offset": offset,
	})

	var templates []*models.NotificationTemplate
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.NotificationTemplate{}).Count(&total).Error; err != nil {
		r.logger.Error(logging.Postgres, logging.Select, err.Error(), nil)
		return nil, 0, err
	}

	// Get paginated results
	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&templates)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), nil)
		return nil, 0, result.Error
	}

	return templates, total, nil
}

func (r *notificationTemplateRepository) GetByType(ctx context.Context, notifType models.NotificationType) ([]*models.NotificationTemplate, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting templates by type", map[logging.ExtraKey]interface{}{
		"type": notifType,
	})

	var templates []*models.NotificationTemplate
	result := r.db.WithContext(ctx).
		Where("type = ?", notifType).
		Order("name ASC").
		Find(&templates)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"type": notifType,
		})
		return nil, result.Error
	}

	return templates, nil
}

func (r *notificationTemplateRepository) Update(ctx context.Context, template *models.NotificationTemplate) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Updating template", map[logging.ExtraKey]interface{}{
		"id": template.ID,
	})

	result := r.db.WithContext(ctx).Save(template)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": template.ID,
		})
		return result.Error
	}

	return nil
}

func (r *notificationTemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug(logging.Postgres, logging.Delete, "Deleting template", map[logging.ExtraKey]interface{}{
		"id": id,
	})

	result := r.db.WithContext(ctx).Delete(&models.NotificationTemplate{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Delete, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}

	return nil
}

func (r *notificationTemplateRepository) GetByKeyAndLocale(ctx context.Context, key string, locale string) (*models.NotificationTemplate, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting template by key and locale", map[logging.ExtraKey]interface{}{
		"key":    key,
		"locale": locale,
	})

	var template models.NotificationTemplate

	// Try exact locale match first
	result := r.db.WithContext(ctx).
		Where("key = ? AND locale = ? AND is_active = ?", key, locale, true).
		First(&template)

	if result.Error == nil {
		return &template, nil
	}

	// If not found and locale is not 'en', fall back to 'en'
	if locale != "en" {
		result = r.db.WithContext(ctx).
			Where("key = ? AND locale = ? AND is_active = ?", key, "en", true).
			First(&template)

		if result.Error == nil {
			return &template, nil
		}
	}

	// No template found
	return nil, result.Error
}

func (r *notificationTemplateRepository) GetActive(ctx context.Context) ([]*models.NotificationTemplate, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting active templates", nil)

	var templates []*models.NotificationTemplate
	result := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&templates)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), nil)
		return nil, result.Error
	}

	return templates, nil
}
