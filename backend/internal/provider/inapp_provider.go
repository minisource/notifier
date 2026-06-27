package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

// InAppProvider delivers in-app notifications by persisting them directly to the database.
// It implements the unified Provider interface.
type InAppProvider struct {
	notifRepo repository.NotificationRepository
	logger    logging.Logger
}

// NewInAppProvider creates a new in-app notification provider
func NewInAppProvider(notifRepo repository.NotificationRepository, logger logging.Logger) *InAppProvider {
	return &InAppProvider{
		notifRepo: notifRepo,
		logger:    logger,
	}
}

// Channel returns the in-app channel
func (p *InAppProvider) Channel() Channel {
	return ChannelInApp
}

// Name returns the provider name
func (p *InAppProvider) Name() string {
	return "inapp_db"
}

// Send persists an in-app notification to the database.
// The notification is immediately available for the user to fetch via the notifications API.
func (p *InAppProvider) Send(ctx context.Context, msg *Message) (*SendResult, error) {
	// Parse user ID
	userID, err := uuid.Parse(msg.UserID)
	if err != nil {
		return NewFailureResult(p.Name(), p.Channel(), ErrorInvalidRecipient, fmt.Sprintf("invalid user ID: %s", msg.UserID)), nil
	}

	now := time.Now()
	notification := &models.Notification{
		UserID:      userID,
		Type:        models.NotificationTypeInApp,
		Status:      models.NotificationStatusSent,
		Subject:     msg.Subject,
		Body:        msg.Body,
		SentAt:      &now,
		DeliveredAt: &now,
		Metadata:    "{}",
	}

	if err := p.notifRepo.Create(ctx, notification); err != nil {
		p.logger.Error(logging.Postgres, logging.Insert, "Failed to create in-app notification", map[logging.ExtraKey]interface{}{
			"userId": userID,
			"error":  err.Error(),
		})
		return NewFailureResult(p.Name(), p.Channel(), ErrorProviderError, "failed to persist in-app notification"), nil
	}

	p.logger.Debug(logging.Postgres, logging.Insert, "In-app notification created", map[logging.ExtraKey]interface{}{
		"notificationId": notification.ID,
		"userId":         userID,
	})

	return NewSuccessResult(p.Name(), p.Channel(), notification.ID.String()), nil
}
