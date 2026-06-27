package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"gorm.io/gorm"
)

type ReminderRepository interface {
	Create(ctx context.Context, reminder *models.Reminder) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Reminder, error)
	List(ctx context.Context, filter ReminderListFilter) ([]*models.Reminder, int64, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Reminder, int64, error)
	Update(ctx context.Context, reminder *models.Reminder) error
	Cancel(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindDue(ctx context.Context, limit int) ([]*models.Reminder, error)
	MarkProcessing(ctx context.Context, id uuid.UUID) error
	MarkSent(ctx context.Context, id uuid.UUID, notificationID uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errorMsg string) error
}

type ReminderListFilter struct {
	Page          int
	PageSize      int
	Status        *models.ReminderStatus
	UserID        *uuid.UUID
	TenantID      *uuid.UUID
	ScheduledFrom *time.Time
	ScheduledTo   *time.Time
	Search        string
	SortBy        string
	SortDirection string
}

type reminderRepository struct {
	db     *gorm.DB
	logger logging.Logger
}

func NewReminderRepository(db *gorm.DB, logger logging.Logger) ReminderRepository {
	return &reminderRepository{
		db:     db,
		logger: logger,
	}
}

func (r *reminderRepository) Create(ctx context.Context, reminder *models.Reminder) error {
	r.logger.Debug(logging.Postgres, logging.Insert, "Creating reminder", map[logging.ExtraKey]interface{}{
		"userId": reminder.UserID,
	})
	result := r.db.WithContext(ctx).Create(reminder)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Insert, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"userId": reminder.UserID,
		})
		return result.Error
	}
	return nil
}

func (r *reminderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Reminder, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Getting reminder by ID", map[logging.ExtraKey]interface{}{
		"id": id,
	})
	var reminder models.Reminder
	result := r.db.WithContext(ctx).First(&reminder, "id = ?", id)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Select, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return nil, result.Error
	}
	return &reminder, nil
}

func (r *reminderRepository) List(ctx context.Context, filter ReminderListFilter) ([]*models.Reminder, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Reminder{})

	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.TenantID != nil {
		query = query.Where("tenant_id = ?", *filter.TenantID)
	}
	if filter.ScheduledFrom != nil {
		query = query.Where("scheduled_at >= ?", *filter.ScheduledFrom)
	}
	if filter.ScheduledTo != nil {
		query = query.Where("scheduled_at <= ?", *filter.ScheduledTo)
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("subject ILIKE ? OR body ILIKE ? OR template_key ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	allowedSortFields := map[string]bool{
		"scheduled_at": true, "created_at": true, "updated_at": true, "status": true,
	}
	sortBy := "scheduled_at"
	if filter.SortBy != "" && allowedSortFields[filter.SortBy] {
		sortBy = filter.SortBy
	}
	sortDir := "ASC"
	if filter.SortDirection == "desc" {
		sortDir = "DESC"
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var reminders []*models.Reminder
	result := query.Order(sortBy+" "+sortDir).Limit(pageSize).Offset(offset).Find(&reminders)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	return reminders, total, nil
}

func (r *reminderRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Reminder, int64, error) {
	var reminders []*models.Reminder
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Reminder{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	result := query.Order("scheduled_at ASC").Limit(limit).Offset(offset).Find(&reminders)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	return reminders, total, nil
}

func (r *reminderRepository) Update(ctx context.Context, reminder *models.Reminder) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Updating reminder", map[logging.ExtraKey]interface{}{
		"id": reminder.ID,
	})
	result := r.db.WithContext(ctx).Save(reminder)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": reminder.ID,
		})
		return result.Error
	}
	return nil
}

func (r *reminderRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug(logging.Postgres, logging.Update, "Cancelling reminder", map[logging.ExtraKey]interface{}{
		"id": id,
	})
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.Reminder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       models.ReminderStatusCancelled,
			"cancelled_at": now,
		})
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Update, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}
	return nil
}

func (r *reminderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug(logging.Postgres, logging.Delete, "Deleting reminder", map[logging.ExtraKey]interface{}{
		"id": id,
	})
	result := r.db.WithContext(ctx).Delete(&models.Reminder{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error(logging.Postgres, logging.Delete, result.Error.Error(), map[logging.ExtraKey]interface{}{
			"id": id,
		})
		return result.Error
	}
	return nil
}

func (r *reminderRepository) FindDue(ctx context.Context, limit int) ([]*models.Reminder, error) {
	r.logger.Debug(logging.Postgres, logging.Select, "Finding due reminders", map[logging.ExtraKey]interface{}{
		"limit": limit,
	})
	var reminders []*models.Reminder
	result := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", models.ReminderStatusPending, time.Now()).
		Order("scheduled_at ASC").
		Limit(limit).
		Find(&reminders)
	if result.Error != nil {
		return nil, result.Error
	}
	return reminders, nil
}

func (r *reminderRepository) MarkProcessing(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.Reminder{}).
		Where("id = ? AND status = ?", id, models.ReminderStatusPending).
		Update("status", models.ReminderStatusProcessing).Error
}

func (r *reminderRepository) MarkSent(ctx context.Context, id uuid.UUID, notificationID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.Reminder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          models.ReminderStatusSent,
			"notification_id": notificationID,
			"sent_at":         now,
		}).Error
}

func (r *reminderRepository) MarkFailed(ctx context.Context, id uuid.UUID, errorMsg string) error {
	return r.db.WithContext(ctx).Model(&models.Reminder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     models.ReminderStatusFailed,
			"last_error": errorMsg,
		}).Error
}
