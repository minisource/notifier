package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Providers(router fiber.Router, handler *handlers.ProviderHandler) {
	router.Get("/", handler.ListProviders)
	router.Post("/", handler.CreateProvider)
	router.Get("/health", handler.GetProviderHealth)
	router.Get("/:providerId", handler.GetProvider)
	router.Put("/:providerId", handler.UpdateProvider)
	router.Delete("/:providerId", handler.DeleteProvider)
	router.Patch("/:providerId/status", handler.ToggleProviderStatus)
	router.Patch("/:providerId/default", handler.SetDefaultProvider)
	router.Post("/:providerId/test", handler.TestProvider)
}
