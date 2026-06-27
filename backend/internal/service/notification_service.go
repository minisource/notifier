package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/websocket"
	"github.com/minisource/notifier/internal/worker"
)

type NotificationService struct {
	logger           logging.Logger
	cfg              *config.Config
	notifRepo        repository.NotificationRepository
	templateRepo     repository.NotificationTemplateRepository
	prefRepo         repository.NotificationPreferenceRepository
	logRepo          repository.NotificationLogRepository
	settingRepo      repository.SettingRepository
	smsTemplateRepo  repository.SMSTemplateRepository
	worker           *worker.NotificationWorker
	wsHub            *websocket.Hub
	preferenceFilter *PreferenceFilter
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
		logger:           logger,
		cfg:              cfg,
		notifRepo:        notifRepo,
		templateRepo:     templateRepo,
		prefRepo:         prefRepo,
		logRepo:          logRepo,
		settingRepo:      settingRepo,
		smsTemplateRepo:  smsTemplateRepo,
		worker:           worker,
		wsHub:            wsHub,
		preferenceFilter: NewPreferenceFilter(prefRepo, logger),
	}
}

// CreateNotification creates and enqueues a notification
func (s *NotificationService) CreateNotification(ctx context.Context, notification *models.Notification) error {
	s.logger.Debug(logging.General, logging.Insert, "Creating notification", map[logging.ExtraKey]interface{}{
		"userId": notification.UserID,
		"type":   notification.Type,
	})

	// Check user preferences with enhanced filtering (quiet hours, categories, system bypass)
	category := ParseCategoryFromMetadata(notification.Metadata)
	result, prefErr := s.preferenceFilter.CheckPreference(ctx, notification.UserID, notification.Type, notification.Priority, category)
	if prefErr != nil {
		s.logger.Warn(logging.General, logging.Insert, "Preference check failed, continuing with defaults", map[logging.ExtraKey]interface{}{
			"userId": notification.UserID,
			"type":   notification.Type,
			"error":  prefErr.Error(),
		})
	} else if !result.Allowed {
		s.logger.Info(logging.General, logging.Insert, "Notification blocked by user preference", map[logging.ExtraKey]interface{}{
			"userId":   notification.UserID,
			"type":     notification.Type,
			"reason":   result.Reason,
			"priority": notification.Priority,
		})
		return errors.New("notification blocked by user preference: " + result.Reason)
	} else if result.Reason != "" {
		// Notification allowed but with a note (e.g., quiet hours, digest only)
		s.logger.Info(logging.General, logging.Insert, "Notification allowed with preference note", map[logging.ExtraKey]interface{}{
			"userId":   notification.UserID,
			"type":     notification.Type,
			"reason":   result.Reason,
			"priority": notification.Priority,
		})
		// Store preference bypass info in metadata for later reference
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

	// Handle digest-only preference
	isDigest := false
	if result != nil && result.Reason == "digest_only" {
		isDigest = true
	}

	// Set defaults
	if notification.Status == "" {
		if isDigest {
			notification.Status = models.NotificationStatusDigested
		} else {
			notification.Status = models.NotificationStatusPending
		}
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
		"isDigest":       isDigest,
	})

	// For digest-only notifications, skip enqueue — they accumulate for batch delivery
	if isDigest {
		s.logger.Debug(logging.General, logging.Insert, "Notification queued for digest, skipping immediate enqueue", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"userId":         notification.UserID,
			"type":           notification.Type,
		})
		return nil
	}

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

	// Check user preferences with enhanced filtering (quiet hours, categories, system bypass)
	category := ParseCategoryFromMetadata(notification.Metadata)
	result, prefErr := s.preferenceFilter.CheckPreference(ctx, notification.UserID, notification.Type, notification.Priority, category)
	if prefErr != nil {
		s.logger.Warn(logging.General, logging.Insert, "Preference check failed, continuing with defaults", map[logging.ExtraKey]interface{}{
			"userId": notification.UserID,
			"type":   notification.Type,
			"error":  prefErr.Error(),
		})
	} else if !result.Allowed {
		s.logger.Info(logging.General, logging.Insert, "Notification blocked by user preference", map[logging.ExtraKey]interface{}{
			"userId":   notification.UserID,
			"type":     notification.Type,
			"reason":   result.Reason,
			"priority": notification.Priority,
		})
		return errors.New("notification blocked by user preference: " + result.Reason)
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

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	s.logger.Debug(logging.General, logging.Select, "Getting unread notification count", map[logging.ExtraKey]interface{}{
		"userId": userID,
	})

	return s.notifRepo.GetUnreadCountByUserID(ctx, userID)
}

// MarkAllAsRead marks all unread notifications as read for a user, returns count of updated rows
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	s.logger.Debug(logging.General, logging.Update, "Marking all notifications as read", map[logging.ExtraKey]interface{}{
		"userId": userID,
	})

	return s.notifRepo.MarkAllAsReadByUserID(ctx, userID)
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

// ListAllNotifications retrieves paginated notifications with admin-level filters
func (s *NotificationService) ListAllNotifications(ctx context.Context, filter repository.NotificationListFilter) ([]*models.Notification, int64, error) {
	return s.notifRepo.ListAll(ctx, filter)
}

// RetryNotification retries a failed/dead notification
func (s *NotificationService) RetryNotification(ctx context.Context, notificationID uuid.UUID) error {
	notif, err := s.notifRepo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	// Validate state: only failed, dead, or retrying (if retries remaining)
	switch notif.Status {
	case models.NotificationStatusFailed, models.NotificationStatusDead:
		// Valid for retry
	case models.NotificationStatusRetrying:
		if notif.RetryCount >= notif.MaxRetries {
			return fmt.Errorf("notification has exceeded max retries (%d/%d)", notif.RetryCount, notif.MaxRetries)
		}
		// Allow retry if still retrying but has remaining attempts
	default:
		return fmt.Errorf("cannot retry notification with status '%s': only failed/dead/retrying statuses can be retried", notif.Status)
	}

	// Check max retries
	if notif.RetryCount >= notif.MaxRetries {
		return fmt.Errorf("notification has exceeded max retries (%d/%d)", notif.RetryCount, notif.MaxRetries)
	}

	// Reset status to pending and increment retry count
	now := time.Now()
	notif.Status = models.NotificationStatusPending
	notif.RetryCount++
	notif.UpdatedAt = now

	return s.notifRepo.Update(ctx, notif)
}

// CancelNotification cancels a pending/queued/retrying notification
func (s *NotificationService) CancelNotification(ctx context.Context, notificationID uuid.UUID) error {
	notif, err := s.notifRepo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	// Validate state: only pending, queued, retrying can be cancelled
	switch notif.Status {
	case models.NotificationStatusPending, models.NotificationStatusQueued, models.NotificationStatusRetrying:
		// Valid for cancel
	case models.NotificationStatusProcessing:
		// Allow cancel if processing but not sent yet
	default:
		return fmt.Errorf("cannot cancel notification with status '%s': only pending/queued/retrying statuses can be cancelled", notif.Status)
	}

	now := time.Now()
	notif.Status = models.NotificationStatusCancelled
	notif.CancelledAt = &now
	notif.UpdatedAt = now

	return s.notifRepo.Update(ctx, notif)
}

// MarkAsSeen marks a notification as seen
func (s *NotificationService) MarkAsSeen(ctx context.Context, notificationID uuid.UUID) error {
	return s.notifRepo.MarkAsSeen(ctx, notificationID)
}

// MarkAsClicked marks a notification as clicked
func (s *NotificationService) MarkAsClicked(ctx context.Context, notificationID uuid.UUID) error {
	return s.notifRepo.MarkAsClicked(ctx, notificationID)
}

// GetAttemptsFromLogs maps notification logs to attempt-style responses
func (s *NotificationService) GetAttemptsFromLogs(ctx context.Context, notificationID uuid.UUID) ([]*models.NotificationLog, error) {
	return s.logRepo.GetByNotificationID(ctx, notificationID)
}	// GetPushConfig retrieves push provider config from database
func (s *NotificationService) GetPushConfig(ctx context.Context) (map[string]string, error) {
	setting, err := s.settingRepo.GetByKey(ctx, SettingKeyPushProviders)
	if err != nil {
		return nil, fmt.Errorf("push provider config not found in database")
	}

	// Parse the JSON config value to extract the provider name
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(setting.Value), &cfg); err != nil {
		return nil, fmt.Errorf("invalid push provider config JSON: %w", err)
	}

	providerName := ""
	if p, ok := cfg["provider"].(string); ok {
		providerName = p
	}

	return map[string]string{
		"provider": providerName,
	}, nil
}

// GetByIDempotencyKey finds a notification by idempotency key
func (s *NotificationService) GetByIDempotencyKey(ctx context.Context, key string) (*models.Notification, error) {
	return s.notifRepo.GetByIDempotencyKey(ctx, key)
}

// GetRepository returns the underlying notification repository for direct access
func (s *NotificationService) GetRepository() repository.NotificationRepository {
	return s.notifRepo
}

// MaskEmail masks an email address for PII protection
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***"
	}
	local := parts[0]
	domain := parts[1]
	if len(local) <= 2 {
		return local[:1] + "***@" + domain
	}
	return local[:1] + "***" + local[len(local)-1:] + "@" + domain
}

// MaskPhone masks a phone number for PII protection
func MaskPhone(phone string) string {
	if phone == "" || len(phone) < 4 {
		return phone
	}
	masked := phone[:2] + strings.Repeat("*", len(phone)-4) + phone[len(phone)-2:]
	return masked
}

// MaskRecipient returns a masked recipient string based on type heuristic
func MaskRecipient(email, phone, userID string) string {
	if email != "" {
		return MaskEmail(email)
	}
	if phone != "" {
		return MaskPhone(phone)
	}
	if userID != "" {
		if len(userID) > 8 {
			return userID[:8] + "***"
		}
		return userID
	}
	return ""
}
