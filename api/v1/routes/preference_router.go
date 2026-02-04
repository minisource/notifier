package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Preferences(router fiber.Router, handler *handlers.PreferenceHandler) {
	router.Get("/user/:userId", handler.GetUserPreferences)
	router.Put("/user/:userId", handler.UpdatePreference)
}
