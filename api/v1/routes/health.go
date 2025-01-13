package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Health(router fiber.Router) {
	handler := handlers.NewHealthHandler()

	router.Get("/", handler.Health)
}