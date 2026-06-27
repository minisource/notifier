package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Admin(router fiber.Router, handler *handlers.AdminHandler) {
	// Notifications
	router.Get("/notifications", handler.ListAllNotifications)
	router.Post("/notifications", handler.CreateNotification)
	router.Post("/notifications/read-all", handler.ReadAllNotifications)
	router.Get("/notifications/:notificationId", handler.GetNotificationByID)
	router.Post("/notifications/:notificationId/retry", handler.RetryNotification)
	router.Post("/notifications/:notificationId/cancel", handler.CancelNotification)
	router.Get("/notifications/:notificationId/attempts", handler.GetAttempts)
	router.Get("/notifications/:notificationId/deliveries", handler.GetDeliveries)

	// Templates
	router.Get("/templates", handler.GetAllTemplates)
	router.Post("/templates", handler.CreateTemplate)
	router.Get("/templates/key/:key", handler.GetTemplateByKey)
	router.Post("/templates/render-preview", handler.RenderPreviewByKey)
	router.Get("/templates/:templateId", handler.GetTemplate)
	router.Put("/templates/:templateId", handler.UpdateTemplate)
	router.Delete("/templates/:templateId", handler.DeleteTemplate)
	router.Post("/templates/:templateId/render-preview", handler.RenderPreview)
	router.Patch("/templates/:templateId/status", handler.UpdateTemplateStatus)

	// Preferences
	router.Get("/preferences/user/:userId", handler.GetUserPreferences)
	router.Put("/preferences/user/:userId", handler.UpdatePreference)
	router.Patch("/preferences/user/:userId/channel/:channel", handler.UpdateChannelPreference)
	router.Patch("/preferences/user/:userId/category/:category", handler.UpdateCategoryPreference)

	// Reminders
	router.Get("/reminders", handler.ListAllReminders)
	router.Post("/reminders", handler.CreateReminder)
	router.Get("/reminders/:reminderId", handler.GetReminder)
	router.Put("/reminders/:reminderId", handler.UpdateReminder)
	router.Post("/reminders/:reminderId/cancel", handler.CancelReminder)
	router.Delete("/reminders/:reminderId", handler.DeleteReminder)
	router.Get("/reminders/user/:userId", handler.GetUserReminders)

	// Admin notification actions
	router.Put("/notifications/:notificationId/read", handler.MarkAsRead)
	router.Post("/notifications/:notificationId/seen", handler.MarkAsSeen)
	router.Post("/notifications/:notificationId/click", handler.MarkAsClicked)

	// Providers — not on AdminHandler; registered in api.go via ProviderHandler under /admin/providers
}
