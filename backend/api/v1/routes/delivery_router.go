package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Deliveries(router fiber.Router, handler *handlers.DeliveryHandler) {
	router.Get("/", handler.ListDeliveries)
	router.Get("/:deliveryId", handler.GetDelivery)
	router.Post("/:deliveryId/retry", handler.RetryDelivery)
}
