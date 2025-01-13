package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/api/v1/handlers"
	"github.com/minisource/notifier/config"
)

func SMS(router fiber.Router, cfg *config.Config) {
	h := handlers.NewSMSHandler(cfg)

	router.Post("/", h.SendSMS)
}
