package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
)

// SMSTemplateRepository handles SMS template database operations
type SMSTemplateRepository interface {
	// GetByKeyAndProvider fetches SMS template by key and provider
	// It first checks for tenant-specific template, then falls back to global template
	GetByKeyAndProvider(ctx context.Context, key, provider string, tenantID *uuid.UUID) (*models.SMSTemplate, error)

	// Create creates a new SMS template
	Create(ctx context.Context, template *models.SMSTemplate) error

	// Update updates an existing SMS template
	Update(ctx context.Context, template *models.SMSTemplate) error

	// Delete deletes an SMS template by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID fetches SMS template by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.SMSTemplate, error)

	// GetAll fetches all SMS templates with pagination
	GetAll(ctx context.Context, limit, offset int) ([]*models.SMSTemplate, int64, error)

	// GetByProvider fetches all SMS templates for a specific provider
	GetByProvider(ctx context.Context, provider string) ([]*models.SMSTemplate, error)
}

type smsTemplateRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

// NewSMSTemplateRepository creates a new SMS template repository
func NewSMSTemplateRepository(db *gorm.DB, logger logging.Logger) SMSTemplateRepository {
	return &smsTemplateRepository{
		db:     db,
		logger: logger,
	}
}

// GetByKeyAndProvider fetches SMS template by key and provider
// Priority: tenant-specific template > global template (tenant_id IS NULL)
func (r *smsTemplateRepository) GetByKeyAndProvider(ctx context.Context, key, provider string, tenantID *uuid.UUID) (*models.SMSTemplate, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting SMS template by key and provider", map[logging.ExtraKey]interface{}{
		"key":      key,
		"provider": provider,
		"tenantId": tenantID,
	})

	var template models.SMSTemplate

	// First, try to find tenant-specific template
	if tenantID != nil {
		result := r.db.WithContext(ctx).
			Where("key = ? AND provider = ? AND tenant_id = ?", key, provider, *tenantID).
			First(&template)

		if result.Error == nil {
			r.logger.Debug(logging.Postgres, logging.Select, "Found tenant-specific SMS template", map[logging.ExtraKey]interface{}{
				"templateId": template.ID,
				"key":        key,
				"provider":   provider,
			})
			return &template, nil
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
				"key":      key,
				"provider": provider,
			})
			return nil, result.Error
		}
	}

	// Fall back to global template (tenant_id IS NULL)
	result := r.db.WithContext(ctx).
		Where("key = ? AND provider = ? AND tenant_id IS NULL", key, provider).
		First(&template)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.Debug(logging.Postgres, logging.Select, "SMS template not found", map[logging.ExtraKey]interface{}{
				"key":      key,
				"provider": provider,
			})
			return nil, result.Error
		}

		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"key":      key,
			"provider": provider,
		})
		return nil, result.Error
	}

	r.logger.Debug(logging.Postgres, logging.Select, "Found global SMS template", map[logging.ExtraKey]interface{}{
		"templateId": template.ID,
		"key":        key,
		"provider":   provider,
	})
	return &template, nil
}

// Create creates a new SMS template
func (r *smsTemplateRepository) Create(ctx context.Context, template *models.SMSTemplate) error {
	r.logger.Debug(logging.Postgres, logging.Insert, "Creating SMS template", map[logging.ExtraKey]interface{}{
		"key":      template.Key,
		"provider": template.Provider,
	})

	result := r.db.WithContext(ctx).Create(template)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Insert, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"key":      template.Key,
			"provider": template.Provider,
		})
		return result.Error
	}

	r.logger.Debug(logging.Postgres, logging.Insert, "SMS template created successfully", map[logging.ExtraKey]interface{}{
		"templateId": template.ID,
	})
	return nil
}

// Update updates an existing SMS template
func (r *smsTemplateRepository) Update(ctx context.Context, template *models.SMSTemplate) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Updating SMS template", map[logging.ExtraKey]interface{}{
		"id":  template.ID,
		"key": template.Key,
	})

	result := r.db.WithContext(ctx).Save(template)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": template.ID,
		})
		return result.Error
	}

	r.logger.Debug(logging.Postgres, logging.Update, "SMS template updated successfully", map[logging.ExtraKey]interface{}{
		"templateId": template.ID,
	})
	return nil
}

// Delete deletes an SMS template by ID
func (r *smsTemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug(logging.Postgres, logging.Delete, "Deleting SMS template", map[logging.ExtraKey]interface{}{
		"id": id,
	})

	result := r.db.WithContext(ctx).Delete(&models.SMSTemplate{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Delete, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}

	r.logger.Debug(logging.Postgres, logging.Delete, "SMS template deleted successfully", map[logging.ExtraKey]interface{}{
		"id": id,
	})
	return nil
}

// GetByID fetches SMS template by ID
func (r *smsTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SMSTemplate, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting SMS template by ID", map[logging.ExtraKey]interface{}{
		"id": id,
	})

	var template models.SMSTemplate
	result := r.db.WithContext(ctx).First(&template, "id = ?", id)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return nil, result.Error
	}

	return &template, nil
}

// GetAll fetches all SMS templates with pagination
func (r *smsTemplateRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.SMSTemplate, int64, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting all SMS templates", map[logging.ExtraKey]interface{}{
		"limit":  limit,
		"offset": offset,
	})

	var templates []*models.SMSTemplate
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&models.SMSTemplate{}).Count(&total).Error; err != nil {
		r.logger.Error(logging.Postgres, logging.Select, err.Error(), nil)
		return nil, 0, err
	}

	// Fetch with pagination
	result := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&templates)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), nil)
		return nil, 0, result.Error
	}

	return templates, total, nil
}

// GetByProvider fetches all SMS templates for a specific provider
func (r *smsTemplateRepository) GetByProvider(ctx context.Context, provider string) ([]*models.SMSTemplate, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting SMS templates by provider", map[logging.ExtraKey]interface{}{
		"provider": provider,
	})

	var templates []*models.SMSTemplate
	result := r.db.WithContext(ctx).
		Where("provider = ?", provider).
		Order("key ASC").
		Find(&templates)

	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"provider": provider,
		})
		return nil, result.Error
	}

	return templates, nil
}
