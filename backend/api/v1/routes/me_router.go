package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Me(router fiber.Router, handler *handlers.MeHandler) {
	// Notifications
	router.Get("/notifications", handler.ListMyNotifications)
	router.Get("/notifications/unread", handler.GetMyUnread)
	router.Get("/notifications/unread-count", handler.GetMyUnreadCount)
	router.Post("/notifications/read-all", handler.MarkMyAllRead)
	router.Get("/notifications/:notificationId", handler.GetMyNotification)
	router.Put("/notifications/:notificationId/read", handler.MarkMyNotificationRead)
	router.Post("/notifications/:notificationId/seen", handler.MarkMyNotificationSeen)
	router.Post("/notifications/:notificationId/click", handler.MarkMyNotificationClicked)

	// Preferences
	router.Get("/preferences", handler.GetMyPreferences)
	router.Put("/preferences", handler.UpdateMyPreferences)
	router.Patch("/preferences/channel/:channel", handler.PatchMyChannelPreference)
	router.Patch("/preferences/category/:category", handler.PatchMyCategoryPreference)

	// Reminders
	router.Get("/reminders", handler.ListMyReminders)
	router.Post("/reminders", handler.CreateMyReminder)
	router.Get("/reminders/:reminderId", handler.GetMyReminder)
	router.Put("/reminders/:reminderId", handler.UpdateMyReminder)
	router.Post("/reminders/:reminderId/cancel", handler.CancelMyReminder)
	router.Delete("/reminders/:reminderId", handler.DeleteMyReminder)
}
