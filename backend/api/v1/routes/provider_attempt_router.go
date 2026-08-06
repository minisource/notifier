package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

// ProviderAttempts registers the provider request lifecycle logging routes
// (admin only — the caller must mount them on an admin-authenticated group).
func ProviderAttempts(router fiber.Router, handler *handlers.ProviderAttemptHandler) {
	router.Get("/", handler.ListProviderAttempts)
	router.Get("/:attemptId/events", handler.ListProviderAttemptEvents)
	router.Get("/:attemptId", handler.GetProviderAttempt)
}
