package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/middleware"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Notifications(router fiber.Router, handler *handlers.NotificationHandler) {
	// Admin/global list (auth-protected)
	router.Get("/", handler.ListAllNotifications)

	// User-scoped routes with access control: self or admin only
	userGroup := router.Group("/user", middleware.RequireSelfOrAdminFromParam("userId"))
	userGroup.Get("/:userId", handler.GetUserNotifications)
	userGroup.Get("/:userId/unread", handler.GetUnreadNotifications)
	userGroup.Get("/:userId/unread-count", handler.GetUnreadCount)
	userGroup.Post("/:userId/read-all", handler.MarkAllAsRead)

	// Create routes
	router.Post("/", handler.CreateNotification)
	router.Post("/send", handler.CreateNotification) // Alias for explicit send intent
	router.Post("/batch", handler.CreateBatchNotifications)

	// Delivery/attempt routes (register before :notificationId to avoid param collision)
	router.Get("/:notificationId/deliveries", handler.GetNotificationDeliveries)
	router.Get("/:notificationId/attempts", handler.GetNotificationAttempts)

	// Single notification routes
	router.Get("/:notificationId", handler.GetNotificationByID)
	router.Put("/:notificationId/read", handler.MarkAsRead)

	// Notification action routes
	router.Post("/:notificationId/retry", handler.RetryNotification)
	router.Post("/:notificationId/cancel", handler.CancelNotification)
	router.Post("/:notificationId/seen", handler.MarkAsSeen)
	router.Post("/:notificationId/click", handler.MarkAsClicked)
}
