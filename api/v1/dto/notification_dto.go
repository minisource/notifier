package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/minisource/notifier/internal/models"
)

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID         uuid.UUID                   `json:"userId" validate:"required"`
	Type           models.NotificationType     `json:"type" validate:"required"`
	Priority       models.NotificationPriority `json:"priority"`
	RecipientEmail string                      `json:"recipientEmail"`
	RecipientPhone string                      `json:"recipientPhone"`
	RecipientID    string                      `json:"recipientId"`
	Subject        string                      `json:"subject"`
	Body           string                      `json:"body" validate:"required"`
	Metadata       map[string]interface{}      `json:"metadata"`
	TemplateID     *uuid.UUID                  `json:"templateId"`
	ScheduledAt    *time.Time                  `json:"scheduledAt"`
}

// BatchNotificationRequest represents a batch notification request
type BatchNotificationRequest struct {
	Notifications []CreateNotificationRequest `json:"notifications" validate:"required,min=1,max=100"`
}

// NotificationResponse represents a notification response
type NotificationResponse struct {
	ID             uuid.UUID                   `json:"id"`
	UserID         uuid.UUID                   `json:"userId"`
	Type           models.NotificationType     `json:"type"`
	Status         models.NotificationStatus   `json:"status"`
	Priority       models.NotificationPriority `json:"priority"`
	RecipientEmail string                      `json:"recipientEmail,omitempty"`
	RecipientPhone string                      `json:"recipientPhone,omitempty"`
	Subject        string                      `json:"subject,omitempty"`
	Body           string                      `json:"body"`
	ReadAt         *time.Time                  `json:"readAt,omitempty"`
	SentAt         *time.Time                  `json:"sentAt,omitempty"`
	CreatedAt      time.Time                   `json:"createdAt"`
}

// PaginatedNotificationResponse represents paginated notifications
type PaginatedNotificationResponse struct {
	Data       []*NotificationResponse `json:"data"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"pageSize"`
	TotalPages int                     `json:"totalPages"`
}

// UpdatePreferenceRequest represents a notification preference update
type UpdatePreferenceRequest struct {
	Type             models.NotificationType `json:"type" validate:"required"`
	IsEnabled        bool                    `json:"isEnabled"`
	AllowInstant     bool                    `json:"allowInstant"`
	AllowDigest      bool                    `json:"allowDigest"`
	DigestFrequency  string                  `json:"digestFrequency"`
	CategorySettings map[string]bool         `json:"categorySettings"`
}

// PreferenceResponse represents a notification preference response
type PreferenceResponse struct {
	ID              uuid.UUID               `json:"id"`
	UserID          uuid.UUID               `json:"userId"`
	Type            models.NotificationType `json:"type"`
	IsEnabled       bool                    `json:"isEnabled"`
	AllowInstant    bool                    `json:"allowInstant"`
	AllowDigest     bool                    `json:"allowDigest"`
	DigestFrequency string                  `json:"digestFrequency"`
}

// CreateTemplateRequest represents a template creation request
type CreateTemplateRequest struct {
	Name             string                  `json:"name" validate:"required"`
	Type             models.NotificationType `json:"type" validate:"required"`
	Subject          string                  `json:"subject"`
	Body             string                  `json:"body" validate:"required"`
	Description      string                  `json:"description"`
	Variables        []string                `json:"variables"`
	Provider         string                  `json:"provider"`
	ProviderTemplate string                  `json:"providerTemplate"`
}

// TemplateResponse represents a template response
type TemplateResponse struct {
	ID          uuid.UUID               `json:"id"`
	Name        string                  `json:"name"`
	Type        models.NotificationType `json:"type"`
	Subject     string                  `json:"subject,omitempty"`
	Body        string                  `json:"body"`
	Description string                  `json:"description,omitempty"`
	IsActive    bool                    `json:"isActive"`
	CreatedAt   time.Time               `json:"createdAt"`
}
