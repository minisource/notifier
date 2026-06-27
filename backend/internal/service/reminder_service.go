package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

// ReminderService handles reminder operations
type ReminderService struct {
	reminderRepo repository.ReminderRepository
	notifService *NotificationService
	logger       logging.Logger
}

// NewReminderService creates a new reminder service
func NewReminderService(
	reminderRepo repository.ReminderRepository,
	notifService *NotificationService,
	logger logging.Logger,
) *ReminderService {
	return &ReminderService{
		reminderRepo: reminderRepo,
		notifService: notifService,
		logger:       logger,
	}
}

// CreateReminder creates a new scheduled reminder
func (s *ReminderService) CreateReminder(ctx context.Context, reminder *models.Reminder) error {
	if err := s.validateReminder(reminder); err != nil {
		return err
	}

	if reminder.ID == uuid.Nil {
		reminder.ID = uuid.New()
	}
	if reminder.Status == "" {
		reminder.Status = models.ReminderStatusPending
	}

	return s.reminderRepo.Create(ctx, reminder)
}

// GetReminder retrieves a reminder by ID
func (s *ReminderService) GetReminder(ctx context.Context, id uuid.UUID) (*models.Reminder, error) {
	reminder, err := s.reminderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("reminder not found: %w", err)
	}
	return reminder, nil
}

// ListReminders retrieves paginated reminders with filters
func (s *ReminderService) ListReminders(ctx context.Context, filter repository.ReminderListFilter) ([]*models.Reminder, int64, error) {
	return s.reminderRepo.List(ctx, filter)
}

// ListUserReminders retrieves reminders for a specific user
func (s *ReminderService) ListUserReminders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Reminder, int64, error) {
	return s.reminderRepo.ListByUserID(ctx, userID, limit, offset)
}

// UpdateReminder updates a reminder (only allowed for scheduled status)
func (s *ReminderService) UpdateReminder(ctx context.Context, id uuid.UUID, updates *models.Reminder) (*models.Reminder, error) {
	existing, err := s.reminderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("reminder not found: %w", err)
	}

	if existing.Status != models.ReminderStatusPending {
		return nil, fmt.Errorf("cannot update reminder with status '%s': only scheduled reminders can be updated", existing.Status)
	}

	// Update allowed fields
	existing.ScheduledAt = updates.ScheduledAt
	existing.RecipientEmail = updates.RecipientEmail
	existing.RecipientPhone = updates.RecipientPhone
	existing.TemplateKey = updates.TemplateKey
	existing.Subject = updates.Subject
	existing.Body = updates.Body
	existing.ChannelsJSON = updates.ChannelsJSON
	existing.VariablesJSON = updates.VariablesJSON

	if err := s.reminderRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update reminder: %w", err)
	}

	return existing, nil
}

// DeleteReminder deletes a reminder (only safe for scheduled/cancelled)
func (s *ReminderService) DeleteReminder(ctx context.Context, id uuid.UUID) error {
	existing, err := s.reminderRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	switch existing.Status {
	case models.ReminderStatusPending, models.ReminderStatusCancelled:
		// Can delete
	default:
		return fmt.Errorf("cannot delete reminder with status '%s': only scheduled/cancelled reminders can be deleted", existing.Status)
	}

	return s.reminderRepo.Delete(ctx, id)
}

// CancelReminder cancels a scheduled reminder
func (s *ReminderService) CancelReminder(ctx context.Context, id uuid.UUID) error {
	existing, err := s.reminderRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	switch existing.Status {
	case models.ReminderStatusPending:
		// Can cancel
	case models.ReminderStatusProcessing:
		// Allow cancel if processing but not sent
	default:
		return fmt.Errorf("cannot cancel reminder with status '%s': only scheduled/processing reminders can be cancelled", existing.Status)
	}

	return s.reminderRepo.Cancel(ctx, id)
}

// ProcessDueReminders finds and processes due reminders
func (s *ReminderService) ProcessDueReminders(ctx context.Context, limit int) (int, error) {
	due, err := s.reminderRepo.FindDue(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to find due reminders: %w", err)
	}

	processed := 0
	for _, reminder := range due {
		if err := s.reminderRepo.MarkProcessing(ctx, reminder.ID); err != nil {
			s.logger.Error(logging.General, logging.Update, "Failed to mark reminder processing", map[logging.ExtraKey]interface{}{
				"reminderId": reminder.ID,
				"error":      err.Error(),
			})
			continue
		}

		// Determine channel from reminder's stored channels
		notifType := models.NotificationTypeEmail
		chs := reminder.ParseChannels()
		if len(chs) > 0 {
			notifType = chs[0]
		}

		// Create notification from reminder
		notification := &models.Notification{
			TenantID:       reminder.TenantID,
			UserID:         reminder.UserID,
			Type:           notifType,
			Priority:       models.NotificationPriorityNormal,
			RecipientEmail: reminder.RecipientEmail,
			RecipientPhone: reminder.RecipientPhone,
			Subject:        reminder.Subject,
			Body:           reminder.Body,
			TemplateKey:    reminder.TemplateKey,
		}

		if err := s.notifService.CreateNotification(ctx, notification); err != nil {
			s.logger.Error(logging.General, logging.Api, "Failed to create notification from reminder", map[logging.ExtraKey]interface{}{
				"reminderId": reminder.ID,
				"error":      err.Error(),
			})
			s.reminderRepo.MarkFailed(ctx, reminder.ID, err.Error())
			continue
		}

		if err := s.reminderRepo.MarkSent(ctx, reminder.ID, notification.ID); err != nil {
			s.logger.Error(logging.General, logging.Update, "Failed to mark reminder sent", map[logging.ExtraKey]interface{}{
				"reminderId":     reminder.ID,
				"notificationId": notification.ID,
				"error":          err.Error(),
			})
			continue
		}

		processed++
	}

	return processed, nil
}

// validateReminder validates reminder fields
func (s *ReminderService) validateReminder(reminder *models.Reminder) error {
	if reminder == nil {
		return fmt.Errorf("reminder cannot be nil")
	}
	if reminder.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if reminder.ScheduledAt.Before(time.Now().Add(-time.Minute)) {
		return fmt.Errorf("scheduled time must be in the future")
	}
	if reminder.Body == "" && reminder.TemplateKey == "" {
		return fmt.Errorf("reminder must have either body or template key")
	}
	return nil
}
