package worker

import "errors"

var (
	ErrQueueFull                   = errors.New("notification queue is full")
	ErrUnsupportedNotificationType = errors.New("unsupported notification type")
	ErrNotificationNotFound        = errors.New("notification not found")
	ErrInvalidNotificationData     = errors.New("invalid notification data")
	// ErrDeliveryPaused is returned by send paths when the global outbound
	// delivery pause is active. It is a CONTROL state, never a provider
	// failure: the worker holds the delivery and preserves its retry budget.
	ErrDeliveryPaused = errors.New("outbound delivery is paused")
)
