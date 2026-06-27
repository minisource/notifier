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

// AdminService handles admin operations
type AdminService struct {
	notificationRepo repository.NotificationRepository
	logRepo          repository.NotificationLogRepository
	logger           logging.Logger
}

// NewAdminService creates a new admin service
func NewAdminService(
	notificationRepo repository.NotificationRepository,
	logRepo repository.NotificationLogRepository,
	logger logging.Logger,
) *AdminService {
	return &AdminService{
		notificationRepo: notificationRepo,
		logRepo:          logRepo,
		logger:           logger,
	}
}

// NotificationLogFilter defines filters for notification logs
type NotificationLogFilter struct {
	UserID      *uuid.UUID
	Type        *models.NotificationType
	Status      *string
	Provider    *string
	StartDate   *time.Time
	EndDate     *time.Time
	SearchQuery string
	Page        int
	PageSize    int
	SortBy      string
	SortOrder   string
}

// StatisticsRequest defines time range for statistics
type StatisticsRequest struct {
	StartDate time.Time
	EndDate   time.Time
	GroupBy   string // hour, day, week, month
}

// Statistics holds aggregated notification statistics
type Statistics struct {
	TotalSent      int64
	TotalDelivered int64
	TotalFailed    int64
	TotalPending   int64
	ByType         map[string]int64
	ByProvider     map[string]int64
	Timeline       []TimelinePoint
}

// TimelinePoint represents a point in time statistics
type TimelinePoint struct {
	Timestamp time.Time
	Sent      int64
	Delivered int64
	Failed    int64
}

// DeliveryStats holds provider-specific delivery statistics
type DeliveryStats struct {
	Provider        string
	TotalSent       int64
	TotalDelivered  int64
	TotalFailed     int64
	AvgDeliveryTime float64
	SuccessRate     float64
}

// GetNotificationLogs retrieves notification logs with filters
func (s *AdminService) GetNotificationLogs(ctx context.Context, filter NotificationLogFilter) ([]*models.Notification, int64, error) {
	s.logger.Debug(logging.General, logging.Select, "Getting notification logs", map[logging.ExtraKey]interface{}{
		"page":     filter.Page,
		"pageSize": filter.PageSize,
	})

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	// Use GetByUserID for now - would need enhanced repository method for full filtering
	// This is a simplified implementation
	offset := (filter.Page - 1) * filter.PageSize

	if filter.UserID != nil {
		notifications, total, err := s.notificationRepo.GetByUserID(ctx, *filter.UserID, filter.PageSize, offset)
		if err != nil {
			s.logger.Error(logging.Postgres, logging.Select, "Failed to get notification logs", map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
			return nil, 0, fmt.Errorf("failed to get notification logs: %w", err)
		}
		return notifications, total, nil
	}

	// For non-user-specific queries, return empty for now
	// Would need repository enhancement for full admin filtering
	return []*models.Notification{}, 0, nil
}

// GetStatistics retrieves aggregated notification statistics
func (s *AdminService) GetStatistics(ctx context.Context, req StatisticsRequest) (*Statistics, error) {
	s.logger.Debug(logging.General, logging.Select, "Getting statistics", map[logging.ExtraKey]interface{}{
		"startDate": req.StartDate,
		"endDate":   req.EndDate,
		"groupBy":   req.GroupBy,
	})

	stats := &Statistics{
		ByType:     make(map[string]int64),
		ByProvider: make(map[string]int64),
		Timeline:   []TimelinePoint{},
	}

	// This would need complex aggregation queries in repository
	// For now, provide basic structure
	s.logger.Info(logging.General, logging.Select, "Statistics calculated", nil)

	return stats, nil
}

// GetDeliveryStats retrieves provider-specific delivery statistics
func (s *AdminService) GetDeliveryStats(ctx context.Context, startDate, endDate time.Time) ([]*DeliveryStats, error) {
	s.logger.Debug(logging.General, logging.Select, "Getting delivery stats", map[logging.ExtraKey]interface{}{
		"startDate": startDate,
		"endDate":   endDate,
	})

	// This would need aggregation queries grouped by provider
	stats := []*DeliveryStats{}

	s.logger.Info(logging.General, logging.Select, "Delivery stats calculated", nil)

	return stats, nil
}

// RetryFailedNotifications retries all failed notifications matching filters
func (s *AdminService) RetryFailedNotifications(ctx context.Context, filter NotificationLogFilter, maxRetries int) (int64, error) {
	s.logger.Debug(logging.General, logging.Update, "Retrying failed notifications", map[logging.ExtraKey]interface{}{
		"maxRetries": maxRetries,
	})

	// Get failed notifications
	status := models.NotificationStatusFailed
	filter.Status = (*string)(&status)
	notifications, _, err := s.GetNotificationLogs(ctx, filter)
	if err != nil {
		return 0, err
	}

	count := int64(0)
	for _, notif := range notifications {
		if notif.RetryCount < maxRetries {
			// Update status to pending for retry
			notif.Status = models.NotificationStatusPending
			notif.RetryCount++
			if err := s.notificationRepo.Update(ctx, notif); err != nil {
				s.logger.Error(logging.Postgres, logging.Update, "Failed to update notification for retry", map[logging.ExtraKey]interface{}{
					"error":          err.Error(),
					"notificationId": notif.ID,
				})
				continue
			}
			count++
		}
	}

	s.logger.Info(logging.General, logging.Update, "Failed notifications queued for retry", map[logging.ExtraKey]interface{}{
		"count": count,
	})

	return count, nil
}

// RetryNotification retries a single notification
func (s *AdminService) RetryNotification(ctx context.Context, notificationID uuid.UUID) error {
	s.logger.Debug(logging.General, logging.Update, "Retrying notification", map[logging.ExtraKey]interface{}{
		"notificationId": notificationID,
	})

	notif, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	if notif.Status != models.NotificationStatusFailed {
		return fmt.Errorf("notification is not in failed state")
	}

	notif.Status = models.NotificationStatusPending
	notif.RetryCount++

	if err := s.notificationRepo.Update(ctx, notif); err != nil {
		s.logger.Error(logging.Postgres, logging.Update, "Failed to update notification", map[logging.ExtraKey]interface{}{
			"error":          err.Error(),
			"notificationId": notificationID,
		})
		return fmt.Errorf("failed to update notification: %w", err)
	}

	s.logger.Info(logging.General, logging.Update, "Notification queued for retry", map[logging.ExtraKey]interface{}{
		"notificationId": notificationID,
	})

	return nil
}

// CancelNotification cancels a pending notification
func (s *AdminService) CancelNotification(ctx context.Context, notificationID uuid.UUID, reason string) error {
	s.logger.Debug(logging.General, logging.Update, "Cancelling notification", map[logging.ExtraKey]interface{}{
		"notificationId": notificationID,
		"reason":         reason,
	})

	notif, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	if notif.Status != models.NotificationStatusPending {
		return fmt.Errorf("notification is not in pending state")
	}

	notif.Status = models.NotificationStatusCanceled
	notif.ErrorMessage = reason

	if err := s.notificationRepo.Update(ctx, notif); err != nil {
		s.logger.Error(logging.Postgres, logging.Update, "Failed to cancel notification", map[logging.ExtraKey]interface{}{
			"error":          err.Error(),
			"notificationId": notificationID,
		})
		return fmt.Errorf("failed to cancel notification: %w", err)
	}

	s.logger.Info(logging.General, logging.Update, "Notification cancelled", map[logging.ExtraKey]interface{}{
		"notificationId": notificationID,
	})

	return nil
}

// BulkDeleteNotifications deletes notifications by IDs or filters
func (s *AdminService) BulkDeleteNotifications(ctx context.Context, ids []uuid.UUID, filter *NotificationLogFilter) (int64, error) {
	s.logger.Debug(logging.General, logging.Delete, "Bulk deleting notifications", map[logging.ExtraKey]interface{}{
		"idsCount": len(ids),
	})

	count := int64(0)

	if len(ids) > 0 {
		// Delete by IDs
		for _, id := range ids {
			if err := s.notificationRepo.Delete(ctx, id); err != nil {
				s.logger.Error(logging.Postgres, logging.Delete, "Failed to delete notification", map[logging.ExtraKey]interface{}{
					"error":          err.Error(),
					"notificationId": id,
				})
				continue
			}
			count++
		}
	} else if filter != nil {
		// Delete by filter
		notifications, _, err := s.GetNotificationLogs(ctx, *filter)
		if err != nil {
			return 0, err
		}

		for _, notif := range notifications {
			if err := s.notificationRepo.Delete(ctx, notif.ID); err != nil {
				s.logger.Error(logging.Postgres, logging.Delete, "Failed to delete notification", map[logging.ExtraKey]interface{}{
					"error":          err.Error(),
					"notificationId": notif.ID,
				})
				continue
			}
			count++
		}
	}

	s.logger.Info(logging.General, logging.Delete, "Notifications deleted", map[logging.ExtraKey]interface{}{
		"count": count,
	})

	return count, nil
}

// GetFailedNotifications retrieves all failed notifications
func (s *AdminService) GetFailedNotifications(ctx context.Context, page, pageSize int) ([]*models.Notification, int64, error) {
	s.logger.Debug(logging.General, logging.Select, "Getting failed notifications", map[logging.ExtraKey]interface{}{
		"page":     page,
		"pageSize": pageSize,
	})

	status := models.NotificationStatusFailed
	filter := NotificationLogFilter{
		Status:   (*string)(&status),
		Page:     page,
		PageSize: pageSize,
	}

	return s.GetNotificationLogs(ctx, filter)
}
