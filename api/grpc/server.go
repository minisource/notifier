package grpc

import (
	"fmt"
	"net"
	"time"

	commongrpc "github.com/minisource/go-common/grpc"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-sdk/auth"
	pb "github.com/minisource/go-sdk/notifier/proto/notifier/v1"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/service"
	"google.golang.org/grpc"
)

// Server implements the gRPC notification service
type Server struct {
	pb.UnimplementedNotificationServiceServer
	pb.UnimplementedTemplateServiceServer
	pb.UnimplementedPreferenceServiceServer
	config        *config.Config
	logger        logging.Logger
	notifSvc      *service.NotificationService
	templateSvc   *service.TemplateService
	preferenceSvc *service.PreferenceService
	authClient    *auth.Client
	grpcServer    *grpc.Server
}

// NewServer creates a new gRPC server
func NewServer(
	cfg *config.Config,
	logger logging.Logger,
	notifSvc *service.NotificationService,
	templateSvc *service.TemplateService,
	preferenceSvc *service.PreferenceService,
	authClient *auth.Client,
) *Server {
	return &Server{
		config:        cfg,
		logger:        logger,
		notifSvc:      notifSvc,
		templateSvc:   templateSvc,
		preferenceSvc: preferenceSvc,
		authClient:    authClient,
	}
}

// Start starts the gRPC server
func (s *Server) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create gRPC server with interceptors from go-common
	var opts []grpc.ServerOption
	if s.config.Auth.Enabled && s.authClient != nil {
		// Use common library interceptors with auth client adapter
		validator := auth.NewGRPCClientAdapter(s.authClient)
		interceptorCfg := commongrpc.AuthInterceptorConfig{
			TokenValidator: validator,
			Logger:         s.logger,
			CacheTTL:       5 * time.Minute,
			ScopeMap:       s.getScopeMap(),
			Enabled:        true,
		}
		opts = append(opts, grpc.UnaryInterceptor(commongrpc.UnaryAuthInterceptor(interceptorCfg)))
		opts = append(opts, grpc.StreamInterceptor(commongrpc.StreamAuthInterceptor(interceptorCfg)))
	}

	s.grpcServer = grpc.NewServer(opts...)

	// Register services
	pb.RegisterNotificationServiceServer(s.grpcServer, s)
	pb.RegisterTemplateServiceServer(s.grpcServer, s)
	pb.RegisterPreferenceServiceServer(s.grpcServer, s)

	s.logger.Info(logging.General, logging.Startup, "gRPC server started", map[logging.ExtraKey]interface{}{
		"port": port,
	})

	return s.grpcServer.Serve(lis)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// getScopeMap returns the scope requirements for each gRPC method
func (s *Server) getScopeMap() map[string]string {
	return map[string]string{
		"/notifier.v1.NotificationService/CreateNotification":       "notifications:send",
		"/notifier.v1.NotificationService/CreateBatchNotifications": "notifications:send",
		"/notifier.v1.NotificationService/SendSMS":                  "notifications:send",
		"/notifier.v1.NotificationService/SendEmail":                "notifications:send",
		"/notifier.v1.NotificationService/GetNotification":          "notifications:read",
		"/notifier.v1.NotificationService/GetUserNotifications":     "notifications:read",
		"/notifier.v1.NotificationService/GetUnreadNotifications":   "notifications:read",
		"/notifier.v1.NotificationService/MarkAsRead":               "notifications:update",
		"/notifier.v1.NotificationService/StreamNotifications":      "notifications:read",
		"/notifier.v1.TemplateService/CreateTemplate":               "templates:create",
		"/notifier.v1.TemplateService/UpdateTemplate":               "templates:update",
		"/notifier.v1.TemplateService/DeleteTemplate":               "templates:delete",
		"/notifier.v1.TemplateService/GetTemplate":                  "templates:read",
		"/notifier.v1.TemplateService/GetAllTemplates":              "templates:read",
		"/notifier.v1.PreferenceService/GetUserPreferences":         "preferences:read",
		"/notifier.v1.PreferenceService/UpdatePreference":           "preferences:update",
	}
}
