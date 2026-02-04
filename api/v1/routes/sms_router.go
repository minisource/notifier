package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/config"
)

// SMS routes deprecated - use notification routes with type="sms"
func SMS(router fiber.Router, cfg *config.Config) {
	// SMS endpoints have been migrated to the notification service
	// Use POST /api/v1/notifications with type="sms" instead
}
