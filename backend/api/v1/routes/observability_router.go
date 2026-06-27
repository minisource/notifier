package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Observability(router fiber.Router, handler *handlers.ObservabilityHandler) {
	router.Get("/health", handler.GetHealth)
	router.Get("/readiness", handler.GetReadiness)
	router.Get("/metrics", handler.GetMetrics)
	router.Get("/queue", handler.GetQueueOverview)
	router.Get("/workers", handler.GetWorkersOverview)
}
