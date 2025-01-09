package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/minisource/notifier/api/v1/handlers"
	"github.com/minisource/notifier/config"
)

func SMS(router *gin.RouterGroup, cfg *config.Config) {
	h := handlers.NewSMSHandler(cfg)

	router.POST("/", h.SendSMS)
}
