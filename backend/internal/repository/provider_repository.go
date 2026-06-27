package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
)

type ProviderRepository interface {
	Create(ctx context.Context, provider *models.Provider) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Provider, error)
	List(ctx context.Context, channel string) ([]*models.Provider, error)
	Update(ctx context.Context, provider *models.Provider) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetPrimaryByChannel(ctx context.Context, channel string) (*models.Provider, error)
}

type providerRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

func NewProviderRepository(db *gorm.DB, logger logging.Logger) ProviderRepository {
	return &providerRepository{
		db:     db,
		logger: logger,
	}
}

func (r *providerRepository) Create(ctx context.Context, provider *models.Provider) error {
	r.logger.Debug(logging.Postgres, logging.Insert, "Creating provider", map[logging.ExtraKey]interface{}{
		"name":    provider.Name,
		"channel": provider.Channel,
	})

	result := r.db.WithContext(ctx).Create(provider)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Insert, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"name": provider.Name,
		})
		return result.Error
	}
	return nil
}

func (r *providerRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Provider, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting provider by ID", map[logging.ExtraKey]interface{}{
		"id": id,
	})

	var provider models.Provider
	result := r.db.WithContext(ctx).First(&provider, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return nil, result.Error
	}
	return &provider, nil
}

func (r *providerRepository) List(ctx context.Context, channel string) ([]*models.Provider, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Listing providers", map[logging.ExtraKey]interface{}{
		"channel": channel,
	})

	var providers []*models.Provider
	query := r.db.WithContext(ctx).Order("priority ASC, name ASC")
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	result := query.Find(&providers)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), nil)
		return nil, result.Error
	}
	return providers, nil
}

func (r *providerRepository) Update(ctx context.Context, provider *models.Provider) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Updating provider", map[logging.ExtraKey]interface{}{
		"id":   provider.ID,
		"name": provider.Name,
	})

	result := r.db.WithContext(ctx).Save(provider)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": provider.ID,
		})
		return result.Error
	}
	return nil
}

func (r *providerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug(logging.Postgres, logging.Delete, "Deleting provider", map[logging.ExtraKey]interface{}{
		"id": id,
	})

	result := r.db.WithContext(ctx).Delete(&models.Provider{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Delete, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}
	return nil
}

func (r *providerRepository) GetPrimaryByChannel(ctx context.Context, channel string) (*models.Provider, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting primary provider by channel", map[logging.ExtraKey]interface{}{
		"channel": channel,
	})

	var provider models.Provider
	result := r.db.WithContext(ctx).
		Where("channel = ? AND is_primary = ? AND is_enabled = ?", channel, true, true).
		First(&provider)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"channel": channel,
		})
		return nil, result.Error
	}
	return &provider, nil
}
