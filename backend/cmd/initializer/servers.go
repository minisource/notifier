package initializer

import (
	"github.com/minisource/go-common/logging"
	grpcServer "github.com/minisource/notifier/api/grpc"
	"github.com/minisource/notifier/config"
)

// InitGRPCServer creates and starts gRPC server if enabled
func InitGRPCServer(cfg *config.Config, services *Services, logger logging.Logger) *grpcServer.Server {
	if !cfg.GRPC.Enabled {
		return nil
	}

	grpcSrv := grpcServer.NewServer(
		cfg,
		logger,
		services.Notification,
		services.Template,
		services.Preference,
		services.AuthClient,
	)

	go func() {
		logger.Info(logging.General, logging.Startup, "Starting gRPC server", map[logging.ExtraKey]interface{}{
			"port": cfg.GRPC.Port,
		})
		if err := grpcSrv.Start(cfg.GRPC.Port); err != nil {
			logger.Error(logging.General, logging.Startup, "gRPC server error", map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
		}
	}()

	return grpcSrv
}
