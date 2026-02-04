package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Notifications(router fiber.Router, handler *handlers.NotificationHandler) {
	router.Post("/", handler.CreateNotification)
	router.Post("/batch", handler.CreateBatchNotifications)
	router.Get("/user/:userId", handler.GetUserNotifications)
	router.Get("/user/:userId/unread", handler.GetUnreadNotifications)
	router.Put("/:notificationId/read", handler.MarkAsRead)
}
