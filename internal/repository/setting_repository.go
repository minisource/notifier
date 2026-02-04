package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
)

type SettingRepository interface {
	Create(ctx context.Context, setting *models.Setting) error
	GetByKey(ctx context.Context, key string) (*models.Setting, error)
	GetByCategory(ctx context.Context, category string) ([]*models.Setting, error)
	GetAll(ctx context.Context) ([]*models.Setting, error)
	GetActive(ctx context.Context) ([]*models.Setting, error)
	Update(ctx context.Context, setting *models.Setting) error
	Delete(ctx context.Context, id uuid.UUID) error
	Upsert(ctx context.Context, setting *models.Setting) error
}

type settingRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

func NewSettingRepository(db *gorm.DB, logger logging.Logger) SettingRepository {
	return &settingRepository{
		db:     db,
		logger: logger,
	}
}

func (r *settingRepository) Create(ctx context.Context, setting *models.Setting) error {
	r.logger.Debug(logging.Postgres, logging.Insert, "Creating setting", map[logging.ExtraKey]interface{}{
		"key":      setting.Key,
		"category": setting.Category,
	})

	result := r.db.WithContext(ctx).Create(setting)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Insert, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"key": setting.Key,
		})
		return result.Error
	}

	return nil
}

func (r *settingRepository) GetByKey(ctx context.Context, key string) (*models.Setting, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting setting by key", map[logging.ExtraKey]interface{}{
		"key": key,
	})

	var setting models.Setting
	result := r.db.WithContext(ctx).Where("key = ?", key).First(&setting)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			r.logger.Debug(logging.Postgres, logging.Select, "Setting not found", map[logging.ExtraKey]interface{}{
				"key": key,
			})
		} else {
			r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
				"key": key,
			})
		}
		return nil, result.Error
	}

	return &setting, nil
}

func (r *settingRepository) GetByCategory(ctx context.Context, category string) ([]*models.Setting, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting settings by category", map[logging.ExtraKey]interface{}{
		"category": category,
	})

	var settings []*models.Setting
	result := r.db.WithContext(ctx).
		Where("category = ? AND is_active = ?", category, true).
		Order("key ASC").
		Find(&settings)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"category": category,
		})
		return nil, result.Error
	}

	return settings, nil
}

func (r *settingRepository) GetAll(ctx context.Context) ([]*models.Setting, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting all settings", nil)

	var settings []*models.Setting
	result := r.db.WithContext(ctx).Order("category ASC, key ASC").Find(&settings)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), nil)
		return nil, result.Error
	}

	return settings, nil
}

func (r *settingRepository) GetActive(ctx context.Context) ([]*models.Setting, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting active settings", nil)

	var settings []*models.Setting
	result := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("category ASC, key ASC").
		Find(&settings)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), nil)
		return nil, result.Error
	}

	return settings, nil
}

func (r *settingRepository) Update(ctx context.Context, setting *models.Setting) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Updating setting", map[logging.ExtraKey]interface{}{
		"id":  setting.ID,
		"key": setting.Key,
	})

	result := r.db.WithContext(ctx).Save(setting)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": setting.ID,
		})
		return result.Error
	}

	return nil
}

func (r *settingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug(logging.Postgres, logging.Delete, "Deleting setting", map[logging.ExtraKey]interface{}{
		"id": id,
	})

	result := r.db.WithContext(ctx).Delete(&models.Setting{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Delete, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}

	return nil
}

func (r *settingRepository) Upsert(ctx context.Context, setting *models.Setting) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Upserting setting", map[logging.ExtraKey]interface{}{
		"key": setting.Key,
	})

	var existing models.Setting
	result := r.db.WithContext(ctx).Where("key = ?", setting.Key).First(&existing)

	if result.Error == gorm.ErrRecordNotFound {
		return r.Create(ctx, setting)
	} else if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"key": setting.Key,
		})
		return result.Error
	}

	setting.ID = existing.ID
	return r.Update(ctx, setting)
}
