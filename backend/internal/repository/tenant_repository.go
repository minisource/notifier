package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	List(ctx context.Context, page, pageSize int) ([]models.Tenant, int64, error)
	Update(ctx context.Context, tenant *models.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetStatus(ctx context.Context, id uuid.UUID, isActive bool) error
}

type tenantRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

func NewTenantRepository(db *gorm.DB, logger logging.Logger) TenantRepository {
	return &tenantRepository{db: db, logger: logger}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *tenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	result := r.db.WithContext(ctx).First(&tenant, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &tenant, nil
}

func (r *tenantRepository) GetBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	result := r.db.WithContext(ctx).First(&tenant, "slug = ?", slug)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &tenant, nil
}

func (r *tenantRepository) List(ctx context.Context, page, pageSize int) ([]models.Tenant, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var tenants []models.Tenant
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.Tenant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := r.db.WithContext(ctx).
		Order("is_default DESC, name ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tenants)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	return tenants, total, nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *tenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Tenant{}, "id = ?", id).Error
}

func (r *tenantRepository) SetStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	return r.db.WithContext(ctx).Model(&models.Tenant{}).Where("id = ?", id).Update("is_active", isActive).Error
}
