package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/middleware"
	"github.com/minisource/notifier/api/v1/handlers"
)

func Preferences(router fiber.Router, handler *handlers.PreferenceHandler) {
	// User-scoped routes with access control: self or admin only
	userGroup := router.Group("/user", middleware.RequireSelfOrAdminFromParam("userId"))
	userGroup.Get("/:userId", handler.GetUserPreferences)
	userGroup.Put("/:userId", handler.UpdatePreference)
	userGroup.Patch("/:userId/channel/:channel", handler.UpdateChannelPreference)
	userGroup.Patch("/:userId/category/:category", handler.UpdateCategoryPreference)
}
