package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDConfig configures the Request ID middleware.
type RequestIDConfig struct {
	HeaderName string // Header to read/write (default: X-Request-Id)
	ContextKey string // Context key to store the request ID (default: requestId)
	Generator  func() string
	SkipPaths  []string
}

// DefaultRequestIDConfig returns default configuration.
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		HeaderName: "X-Request-Id",
		ContextKey: "requestId",
		Generator: func() string {
			return uuid.New().String()
		},
	}
}

// RequestIDMiddleware ensures every request has a unique request ID.
// If the incoming request has an X-Request-Id header, it is preserved.
// Otherwise, a new UUID is generated.
// The request ID is stored in c.Locals("requestId") and set on the response header.
func RequestIDMiddleware(cfg ...RequestIDConfig) fiber.Handler {
	config := DefaultRequestIDConfig()
	if len(cfg) > 0 {
		if cfg[0].HeaderName != "" {
			config.HeaderName = cfg[0].HeaderName
		}
		if cfg[0].ContextKey != "" {
			config.ContextKey = cfg[0].ContextKey
		}
		if cfg[0].Generator != nil {
			config.Generator = cfg[0].Generator
		}
		if cfg[0].SkipPaths != nil {
			config.SkipPaths = cfg[0].SkipPaths
		}
	}

	return func(c *fiber.Ctx) error {
		// Skip configured paths
		for _, path := range config.SkipPaths {
			if c.Path() == path {
				return c.Next()
			}
		}

		// Use existing request ID from header, or generate new
		requestID := c.Get(config.HeaderName)
		if requestID == "" {
			requestID = config.Generator()
		}

		// Store in Fiber context (accessible by handlers)
		c.Locals(config.ContextKey, requestID)

		// Set response header
		c.Set(config.HeaderName, requestID)

		return c.Next()
	}
}

// GetRequestID extracts the request ID from the Fiber context.
func GetRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("requestId").(string); ok {
		return id
	}
	return ""
}
