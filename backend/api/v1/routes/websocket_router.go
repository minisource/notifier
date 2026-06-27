package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/notifier/internal/websocket"
)

func WebSocket(router fiber.Router, hub *websocket.Hub) {
	router.Get("/", hub.HandleWebSocket())
}
