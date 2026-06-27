package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Templates(router fiber.Router, handler *handlers.TemplateHandler) {
	// CRUD routes
	router.Post("/", handler.CreateTemplate)
	router.Get("/", handler.GetAllTemplates)

	// Action routes without ID (register before :templateId to avoid param collision)
	router.Post("/render-preview", handler.RenderPreviewByKey)

	// Key-based lookup (register before :templateId to avoid param collision)
	router.Get("/key/:key", handler.GetTemplateByKey)

	// ID-based routes
	router.Get("/:templateId", handler.GetTemplate)
	router.Put("/:templateId", handler.UpdateTemplate)
	router.Delete("/:templateId", handler.DeleteTemplate)
	router.Patch("/:templateId/status", handler.UpdateTemplateStatus)

	// Action routes
	router.Post("/:templateId/render-preview", handler.RenderPreview)
}
