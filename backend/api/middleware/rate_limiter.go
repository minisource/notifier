package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimitConfig configures the rate limiter.
type RateLimitConfig struct {
	Enabled     bool
	Requests    int           // Max requests per window
	Window      time.Duration // Time window
	SkipPaths   []string
	ContextKey  string // Locals key for rate limit info
}

// DefaultRateLimitConfig returns sensible defaults.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:    true,
		Requests:   100,
		Window:     time.Minute,
		SkipPaths:  []string{"/health", "/ready", "/metrics", "/swagger"},
		ContextKey: "rateLimit",
	}
}

// per-route override configs
type routeLimitConfig struct {
	requests int
	window   time.Duration
}

// in-memory rate limiter — simple sliding window per IP
type ipRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	config   RateLimitConfig
	routes   map[string]routeLimitConfig // path prefix -> limit config
}

func newIPRateLimiter(config RateLimitConfig) *ipRateLimiter {
	return &ipRateLimiter{
		requests: make(map[string][]time.Time),
		config:   config,
		routes:   make(map[string]routeLimitConfig),
	}
}

// allow checks whether the request from ip to path is permitted. Returns
// (allowed, window) so callers can emit a bounded Retry-After on rejection.
func (l *ipRateLimiter) allow(ip string, path string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	window := l.config.Window
	maxReqs := l.config.Requests

	// Check per-route override
	for prefix, rlc := range l.routes {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			window = rlc.window
			maxReqs = rlc.requests
			break
		}
	}

	// Clean old entries
	cutoff := now.Add(-window)
	times := l.requests[ip]
	var valid []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= maxReqs {
		l.requests[ip] = valid
		return false, window
	}

	valid = append(valid, now)
	l.requests[ip] = valid
	return true, window
}

// SetRouteLimit sets a per-route override for rate limiting.
// Path prefix matching is used (e.g., "/api/v1/providers" matches all provider routes).
func (l *ipRateLimiter) SetRouteLimit(pathPrefix string, requests int, window time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.routes[pathPrefix] = routeLimitConfig{requests: requests, window: window}
}

// Handler returns a Fiber middleware handler bound to this limiter instance.
// Used when route-specific limits must be registered BEFORE app.Use (see
// api.go: the delivery-control endpoints get strict low-frequency limits).
func (l *ipRateLimiter) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !l.config.Enabled {
			return c.Next()
		}

		// Skip specific paths
		for _, path := range l.config.SkipPaths {
			if c.Path() == path {
				return c.Next()
			}
		}

		ip := c.IP()
		allowed, window := l.allow(ip, c.Path())
		if !allowed {
			// Bounded Retry-After hint so clients do not hammer immediately.
			retryAfter := int(window.Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Set("Retry-After", strconv.Itoa(retryAfter))
			c.Status(fiber.StatusTooManyRequests)
			return c.JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "RATE_LIMITED",
					"message": "Too many requests. Please try again later.",
				},
			})
		}

		return c.Next()
	}
}

// RateLimiterMiddleware returns a Fiber middleware that rate-limits by IP address.
// Uses in-memory sliding window. For production with multiple instances, use Redis-backed limiter.
func RateLimiterMiddleware(cfg ...RateLimitConfig) fiber.Handler {
	config := DefaultRateLimitConfig()
	if len(cfg) > 0 {
		if cfg[0].Requests > 0 {
			config.Requests = cfg[0].Requests
		}
		if cfg[0].Window > 0 {
			config.Window = cfg[0].Window
		}
		config.Enabled = cfg[0].Enabled
		if cfg[0].SkipPaths != nil {
			config.SkipPaths = cfg[0].SkipPaths
		}
	}

	limiter := newIPRateLimiter(config)
	return limiter.Handler()
}

// NewRateLimiter creates a rate limiter instance for programmatic use (e.g., registering route-specific limits).
func NewRateLimiter(config RateLimitConfig) *ipRateLimiter {
	return newIPRateLimiter(config)
}
