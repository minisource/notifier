package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

// ProviderBalance registers admin provider balance/quota monitoring routes.
// The static /balance sub-routes MUST be registered before the :providerId
// param route in the Providers router (they are, because ProviderBalance is
// mounted on the /providers/balance group in api.go).
func ProviderBalance(router fiber.Router, handler *handlers.BalanceHandler) {
	router.Get("/", handler.ListHealth)
	router.Get("/alerts", handler.ListAlerts)
	router.Post("/alerts/:alertId/acknowledge", handler.Acknowledge)
}
