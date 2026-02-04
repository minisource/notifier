package worker

import "errors"

var (
	ErrQueueFull                   = errors.New("notification queue is full")
	ErrUnsupportedNotificationType = errors.New("unsupported notification type")
	ErrNotificationNotFound        = errors.New("notification not found")
	ErrInvalidNotificationData     = errors.New("invalid notification data")
)
