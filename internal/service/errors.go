package service

import (
	"github.com/minisource/go-common/service_errors"
)

// Notification service specific error codes
const (
	NotificationNotFound       = "notification_not_found"
	NotificationAlreadyRead    = "notification_already_read"
	InvalidNotificationType    = "invalid_notification_type"
	InvalidPriority            = "invalid_priority"
	RecipientRequired          = "recipient_required"
	TemplateNotFound           = "template_not_found"
	PreferenceNotFound         = "preference_not_found"
	MaxRetriesReached          = "max_retries_reached"
	NotificationCreated        = "notification_created"
	NotificationSent           = "notification_sent"
	NotificationFailed         = "notification_failed"
	NotificationMarkedRead     = "notification_marked_read"
	NotificationScheduled      = "notification_scheduled"
	TemplateCreated            = "template_created"
	TemplateUpdated            = "template_updated"
	TemplateDeleted            = "template_deleted"
	PreferenceUpdated          = "preference_updated"
	BatchNotificationCompleted = "batch_notification_completed"
)

// Error constructors with i18n support

func NewNotificationNotFoundError() *service_errors.ServiceError {
	return service_errors.NewServiceError(NotificationNotFound, "Notification not found", "")
}

func NewNotificationAlreadyReadError() *service_errors.ServiceError {
	return service_errors.NewServiceError(NotificationAlreadyRead, "Notification already marked as read", "")
}

func NewInvalidNotificationTypeError() *service_errors.ServiceError {
	return service_errors.NewServiceError(InvalidNotificationType, "Invalid notification type", "")
}

func NewInvalidPriorityError() *service_errors.ServiceError {
	return service_errors.NewServiceError(InvalidPriority, "Invalid priority level", "")
}

func NewRecipientRequiredError() *service_errors.ServiceError {
	return service_errors.NewServiceError(RecipientRequired, "At least one recipient is required", "")
}

func NewTemplateNotFoundError() *service_errors.ServiceError {
	return service_errors.NewServiceError(TemplateNotFound, "Template not found", "")
}

func NewPreferenceNotFoundError() *service_errors.ServiceError {
	return service_errors.NewServiceError(PreferenceNotFound, "Preference not found", "")
}

func NewMaxRetriesReachedError() *service_errors.ServiceError {
	return service_errors.NewServiceError(MaxRetriesReached, "Maximum retry attempts reached", "")
}
