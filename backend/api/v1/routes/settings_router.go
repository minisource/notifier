package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Settings(router fiber.Router, handler *handlers.SettingsHandler) {
	router.Get("/notifications", handler.GetNotificationSettings)
	router.Patch("/notifications", handler.UpdateNotificationSettings)
}
