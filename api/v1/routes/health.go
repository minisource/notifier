package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/minisource/notifire/api/v1/handlers"
)

func Health(r *gin.RouterGroup) {
	handler := handlers.NewHealthHandler()

	r.GET("/", handler.Health)
}
