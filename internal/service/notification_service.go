package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/websocket"
	"github.com/minisource/notifier/internal/worker"
)

type NotificationService struct {
	logger          logging.Logger
	cfg             *config.Config
	notifRepo       repository.NotificationRepository
	templateRepo    repository.NotificationTemplateRepository
	prefRepo        repository.NotificationPreferenceRepository
	logRepo         repository.NotificationLogRepository
	settingRepo     repository.SettingRepository
	smsTemplateRepo repository.SMSTemplateRepository
	worker          *worker.NotificationWorker
	wsHub           *websocket.Hub
}

func NewNotificationService(
	cfg *config.Config,
	logger logging.Logger,
	notifRepo repository.NotificationRepository,
	templateRepo repository.NotificationTemplateRepository,
	prefRepo repository.NotificationPreferenceRepository,
	logRepo repository.NotificationLogRepository,
	settingRepo repository.SettingRepository,
	smsTemplateRepo repository.SMSTemplateRepository,
	worker *worker.NotificationWorker,
	wsHub *websocket.Hub,
) *NotificationService {
	return &NotificationService{
		logger:          logger,
		cfg:             cfg,
		notifRepo:       notifRepo,
		templateRepo:    templateRepo,
		prefRepo:        prefRepo,
		logRepo:         logRepo,
		settingRepo:     settingRepo,
		smsTemplateRepo: smsTemplateRepo,
		worker:          worker,
		wsHub:           wsHub,
	}
}

// CreateNotification creates and enqueues a notification
func (s *NotificationService) CreateNotification(ctx context.Context, notification *models.Notification) error {
	s.logger.Debug(logging.General, logging.Insert, "Creating notification", map[logging.ExtraKey]interface{}{
		"userId": notification.UserID,
		"type":   notification.Type,
	})

	// Check user preferences
	pref, err := s.prefRepo.GetByUserIDAndType(ctx, notification.UserID, notification.Type)
	if err == nil && pref != nil && !pref.IsEnabled {
		s.logger.Info(logging.General, logging.Insert, "Notification disabled by user preference", map[logging.ExtraKey]interface{}{
			"userId": notification.UserID,
			"type":   notification.Type,
		})
		return errors.New("notifications disabled for this type")
	}

	// Apply template if specified
	if notification.TemplateID != nil {
		template, err := s.templateRepo.GetByID(ctx, *notification.TemplateID)
		if err != nil {
			s.logger.Error(logging.General, logging.Select, "Failed to get template", map[logging.ExtraKey]interface{}{
				"templateId": *notification.TemplateID,
				"error":      err.Error(),
			})
			return err
		}

		// Apply template content
		if notification.Subject == "" {
			notification.Subject = template.Subject
		}
		if notification.Body == "" {
			notification.Body = template.Body
		}
		notification.Provider = template.Provider
	}

	// Set defaults
	if notification.Status == "" {
		notification.Status = models.NotificationStatusPending
	}
	if notification.Priority == "" {
		notification.Priority = models.NotificationPriorityNormal
	}
	if notification.MaxRetries == 0 {
		notification.MaxRetries = 3
	}

	// Save to database
	if err := s.notifRepo.Create(ctx, notification); err != nil {
		s.logger.Error(logging.General, logging.Insert, "Failed to create notification", map[logging.ExtraKey]interface{}{
			"userId": notification.UserID,
			"error":  err.Error(),
		})
		return err
	}

	s.logger.Info(logging.General, logging.Insert, "Notification created successfully", map[logging.ExtraKey]interface{}{
		"notificationId": notification.ID,
		"userId":         notification.UserID,
		"type":           notification.Type,
	})

	// Enqueue for processing
	if err := s.worker.EnqueueNotification(notification); err != nil {
		s.logger.Error(logging.Internal, logging.Api, "Failed to enqueue notification", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"error":          err.Error(),
		})
		return err
	}

	// For in-app notifications, broadcast via WebSocket
	if notification.Type == models.NotificationTypeInApp {
		s.wsHub.BroadcastToUser(notification.UserID, notification)
	}

	return nil
}

// CreateNotificationSync creates and sends notification synchronously
// This method waits for the actual send operation to complete and returns real errors
// Use this for critical operations like OTP where you need to know if it actually sent
func (s *NotificationService) CreateNotificationSync(ctx context.Context, notification *models.Notification) error {
	s.logger.Debug(logging.General, logging.Insert, "Creating notification (sync)", map[logging.ExtraKey]interface{}{
		"userId": notification.UserID,
		"type":   notification.Type,
	})

	// Check user preferences
	pref, err := s.prefRepo.GetByUserIDAndType(ctx, notification.UserID, notification.Type)
	if err == nil && pref != nil && !pref.IsEnabled {
		return errors.New("notifications disabled for this type")
	}

	// Set defaults
	if notification.Status == "" {
		notification.Status = models.NotificationStatusPending
	}
	if notification.Priority == "" {
		notification.Priority = models.NotificationPriorityNormal
	}
	if notification.MaxRetries == 0 {
		notification.MaxRetries = 3
	}

	// Save to database
	if err := s.notifRepo.Create(ctx, notification); err != nil {
		s.logger.Error(logging.General, logging.Insert, "Failed to create notification", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Send synchronously using worker's direct send capability
	// The worker has access to all handlers and can send immediately
	notification.Status = models.NotificationStatusSending
	s.notifRepo.UpdateStatus(ctx, notification.ID, models.NotificationStatusSending)

	sendErr := s.worker.SendNotificationSync(ctx, notification)

	if sendErr != nil {
		s.logger.Error(logging.General, logging.Api, "Failed to send notification (sync)", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"error":          sendErr.Error(),
		})
		notification.Status = models.NotificationStatusFailed
		notification.ErrorMessage = sendErr.Error()
		s.notifRepo.UpdateStatus(ctx, notification.ID, models.NotificationStatusFailed)
		return fmt.Errorf("failed to send %s: %w", notification.Type, sendErr)
	}

	s.logger.Info(logging.General, logging.Api, "Notification sent successfully (sync)", map[logging.ExtraKey]interface{}{
		"notificationId": notification.ID,
	})
	notification.Status = models.NotificationStatusSent
	s.notifRepo.UpdateStatus(ctx, notification.ID, models.NotificationStatusSent)
	return nil
}

// CreateBatchNotifications creates multiple notifications
func (s *NotificationService) CreateBatchNotifications(ctx context.Context, notifications []*models.Notification) ([]uuid.UUID, []error) {
	s.logger.Info(logging.General, logging.Insert, "Creating batch notifications", map[logging.ExtraKey]interface{}{
		"count": len(notifications),
	})

	successIDs := make([]uuid.UUID, 0)
	errors := make([]error, 0)

	for _, notification := range notifications {
		if err := s.CreateNotification(ctx, notification); err != nil {
			errors = append(errors, err)
		} else {
			successIDs = append(successIDs, notification.ID)
		}
	}

	s.logger.Info(logging.General, logging.Insert, "Batch notifications created", map[logging.ExtraKey]interface{}{
		"total":   len(notifications),
		"success": len(successIDs),
		"failed":  len(errors),
	})

	return successIDs, errors
}

// GetNotification retrieves a notification by ID
func (s *NotificationService) GetNotification(ctx context.Context, notificationID uuid.UUID) (*models.Notification, error) {
	s.logger.Debug(logging.General, logging.Select, "Getting notification", map[logging.ExtraKey]interface{}{
		"notificationId": notificationID,
	})

	return s.notifRepo.GetByID(ctx, notificationID)
}

// GetUserNotifications retrieves notifications for a user
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Notification, int64, error) {
	s.logger.Debug(logging.General, logging.Select, "Getting user notifications", map[logging.ExtraKey]interface{}{
		"userId": userID,
		"limit":  limit,
		"offset": offset,
	})

	return s.notifRepo.GetByUserID(ctx, userID, limit, offset)
}

// GetUnreadNotifications retrieves unread notifications for a user
func (s *NotificationService) GetUnreadNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Notification, int64, error) {
	s.logger.Debug(logging.General, logging.Select, "Getting unread notifications", map[logging.ExtraKey]interface{}{
		"userId": userID,
	})

	return s.notifRepo.GetUnreadByUserID(ctx, userID, limit, offset)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	s.logger.Debug(logging.General, logging.Update, "Marking notification as read", map[logging.ExtraKey]interface{}{
		"notificationId": notificationID,
	})

	return s.notifRepo.MarkAsRead(ctx, notificationID)
}

// GetNotificationLogs retrieves logs for a notification
func (s *NotificationService) GetNotificationLogs(ctx context.Context, notificationID uuid.UUID) ([]*models.NotificationLog, error) {
	return s.logRepo.GetByNotificationID(ctx, notificationID)
}
