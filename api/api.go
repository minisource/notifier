package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/minisource/common_go/http/middlewares"
	"github.com/minisource/common_go/logging"
	routers "github.com/minisource/notifier/api/v1/routes"
	"github.com/minisource/notifier/config"
)

var logger = logging.NewLogger(&config.GetConfig().Logger)

func InitServer(cfg *config.Config) {
	gin.SetMode(cfg.Server.RunMode)
	r := gin.New()
	// RegisterValidators()

	r.Use(middlewares.DefaultStructuredLogger(&cfg.Logger))
	r.Use(middlewares.Cors(cfg.Cors.AllowOrigins))
	r.Use(gin.Logger(), gin.CustomRecovery(middlewares.ErrorHandler) /*middlewares.TestMiddleware()*/, middlewares.LimitByRequest())

	RegisterRoutes(r, cfg)
	// RegisterSwagger(r, cfg)

	logger := logging.NewLogger(&cfg.Logger)
	logger.Info(logging.General, logging.Startup, "Started", nil)
	err := r.Run(fmt.Sprintf(":%s", cfg.Server.InternalPort))
	if err != nil {
		logger.Fatal(logging.General, logging.Startup, err.Error(), nil)
	}
}

func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	api := r.Group("/api")

	v1 := api.Group("/v1")
	{
		// Test
		health := v1.Group("/health")
		test_router := v1.Group("/test")

		routers.Health(health)
		routers.TestRouter(test_router)

		// sms
		sms := v1.Group("/sms")
		routers.SMS(sms, cfg)
	}
}

// func RegisterSwagger(r *gin.Engine, cfg *config.Config) {
// 	docs.SwaggerInfo.Title = "golang web api"
// 	docs.SwaggerInfo.Description = "golang web api"
// 	docs.SwaggerInfo.Version = "1.0"
// 	docs.SwaggerInfo.BasePath = "/api"
// 	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", cfg.Server.ExternalPort)
// 	docs.SwaggerInfo.Schemes = []string{"http"}

// 	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
// }
