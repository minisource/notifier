package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func InApp(router fiber.Router, handler *handlers.InAppHandler) {
	// In-app notification inbox
	router.Get("/user/:userId", handler.GetInAppNotifications)

	// Read receipt / tracking endpoints
	router.Post("/:id/seen", handler.MarkAsSeen)
	router.Post("/:id/read", handler.MarkAsRead)
	router.Post("/:id/click", handler.MarkAsClicked)
}
