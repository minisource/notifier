package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/logging"
)

// RequestTimingConfig configures the request timing middleware.
type RequestTimingConfig struct {
	// SlowRequestThreshold is the duration after which a request is logged as SLOW_REQUEST.
	// Default: 1 second.
	SlowRequestThreshold time.Duration

	// Logger to use for slow request logging.
	Logger logging.Logger
}

// DefaultRequestTimingConfig returns default configuration.
func DefaultRequestTimingConfig() RequestTimingConfig {
	return RequestTimingConfig{
		SlowRequestThreshold: 1 * time.Second,
	}
}

// RequestTimingMiddleware measures request duration and logs slow requests.
// It adds request duration to the Fiber context for downstream use.
func RequestTimingMiddleware(cfg ...RequestTimingConfig) fiber.Handler {
	config := DefaultRequestTimingConfig()
	if len(cfg) > 0 {
		if cfg[0].SlowRequestThreshold > 0 {
			config.SlowRequestThreshold = cfg[0].SlowRequestThreshold
		}
		if cfg[0].Logger != nil {
			config.Logger = cfg[0].Logger
		}
	}

	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process the request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)
		durationMs := duration.Milliseconds()

		// Store duration in context for downstream use
		c.Locals("requestDuration", duration)
		c.Locals("requestDurationMs", durationMs)

		// Log slow requests
		if duration >= config.SlowRequestThreshold && config.Logger != nil {
			requestID := GetRequestID(c)
			config.Logger.Warn(logging.General, logging.Api, "SLOW_REQUEST", map[logging.ExtraKey]interface{}{
				"requestId": requestID,
				"method":    c.Method(),
				"path":      c.Path(),
				"status":    c.Response().StatusCode(),
				"durationMs": durationMs,
			})
		}

		return err
	}
}

// GetRequestDurationMs returns the request duration in milliseconds from the Fiber context.
func GetRequestDurationMs(c *fiber.Ctx) int64 {
	if d, ok := c.Locals("requestDurationMs").(int64); ok {
		return d
	}
	return 0
}
