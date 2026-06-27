package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/middleware"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Reminders(router fiber.Router, handler *handlers.ReminderHandler) {
	router.Get("/", handler.ListReminders)
	router.Post("/", handler.CreateReminder)

	// User-scoped routes with access control: self or admin only
	userGroup := router.Group("/user", middleware.RequireSelfOrAdminFromParam("userId"))
	userGroup.Get("/:userId", handler.GetUserReminders)

	router.Get("/:reminderId", handler.GetReminder)
	router.Put("/:reminderId", handler.UpdateReminder)
	router.Delete("/:reminderId", handler.DeleteReminder)
	router.Post("/:reminderId/cancel", handler.CancelReminder)
}
