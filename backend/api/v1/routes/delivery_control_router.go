package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

// DeliveryControl registers the admin global outbound delivery pause routes.
func DeliveryControl(router fiber.Router, handler *handlers.DeliveryControlHandler) {
	router.Get("/status", handler.Status)
	router.Post("/pause", handler.Pause)
	router.Post("/resume", handler.Resume)
	router.Get("/history", handler.History)
	router.Get("/held", handler.Held)
}
