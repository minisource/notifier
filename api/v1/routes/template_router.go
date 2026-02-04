package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Templates(router fiber.Router, handler *handlers.TemplateHandler) {
	router.Post("/", handler.CreateTemplate)
	router.Get("/", handler.GetAllTemplates)
	router.Get("/:templateId", handler.GetTemplate)
	router.Put("/:templateId", handler.UpdateTemplate)
	router.Delete("/:templateId", handler.DeleteTemplate)
}
