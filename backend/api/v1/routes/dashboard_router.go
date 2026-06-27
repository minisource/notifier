package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Dashboard(router fiber.Router, handler *handlers.DashboardHandler) {
	router.Get("/overview", handler.GetDashboardOverview)
}
