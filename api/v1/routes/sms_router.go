package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/minisource/notifire/api/v1/handlers"
	"github.com/minisource/notifire/config"
)

func SMS(router *gin.RouterGroup, cfg *config.Config) {
	h := handlers.NewSMSHandler(cfg)

	router.POST("/", h.SendSMS)
}