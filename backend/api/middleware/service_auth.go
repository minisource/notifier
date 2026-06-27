package middleware

// This file re-exports the common service auth middleware from go-common.
// All microservices should use the common library to avoid duplication.

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/http/middleware"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-sdk/auth"
)

// ServiceAuthConfig holds configuration for service auth middleware
// This wraps the auth.Client and provides a simple interface for microservices
type ServiceAuthConfig struct {
	AuthClient    *auth.Client
	Logger        logging.Logger
	CacheTTL      time.Duration
	SkipPaths     []string
	RequiredScope string
	Enabled       bool
}

// ServiceAuthMiddleware creates service authentication middleware using go-common
// This is a convenience wrapper that uses the auth.Client as the token validator
func ServiceAuthMiddleware(cfg ServiceAuthConfig) fiber.Handler {
	// Create adapter for auth client
	validator := auth.NewHTTPClientAdapter(cfg.AuthClient)

	enabled := true
	if cfg.AuthClient == nil {
		enabled = false
	}

	return middleware.RemoteServiceAuthMiddleware(middleware.RemoteServiceAuthConfig{
		TokenValidator: validator,
		Logger:         cfg.Logger,
		CacheTTL:       cfg.CacheTTL,
		SkipPaths:      cfg.SkipPaths,
		RequiredScope:  cfg.RequiredScope,
		Enabled:        enabled,
	})
}

// RequireScope creates a middleware that checks for a specific scope
// Re-exports from go-common
var RequireScope = middleware.RequireScope
