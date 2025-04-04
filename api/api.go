package api

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/minisource/common_go/http/middleware"
	"github.com/minisource/common_go/logging"
	routers "github.com/minisource/notifier/api/v1/routes"
	"github.com/minisource/notifier/config"
)

func InitServer(cfg *config.Config) {
	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		// ErrorHandler: handler.CustomErrorHandler(logger),
		AppName:     cfg.Server.Name,
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	// Middleware
	app.Use(middleware.DefaultStructuredLogger(&cfg.Logger)) // Custom structured logger
	app.Use(middleware.Cors(cfg.Cors.AllowOrigins))
	app.Use(recover.New())
	// app.Use(middleware.LimitByRequest()) // Custom rate limiter

	// Register routes
	RegisterRoutes(app, cfg)

	// Start the server
	logger := logging.NewLogger(&cfg.Logger)
	logger.Info(logging.General, logging.Startup, "Server started", nil)

	err := app.Listen(fmt.Sprintf(":%s", cfg.Server.InternalPort))
	if err != nil {
		logger.Fatal(logging.General, logging.Startup, err.Error(), nil)
	}
}

func RegisterRoutes(app *fiber.App, cfg *config.Config) {
	// Create an API group
	api := app.Group("/api")

	// Create a v1 group
	v1 := api.Group("/v1")
	{
		// Health routes
		health := v1.Group("/health")
		routers.Health(health)

		// SMS routes
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
