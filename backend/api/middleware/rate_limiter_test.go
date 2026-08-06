package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TestRateLimiter_RouteSpecificLimit verifies that a strict route limit
// (delivery-control pause/resume) overrides the global limit and that a 429
// response carries a bounded Retry-After header.
func TestRateLimiter_RouteSpecificLimit(t *testing.T) {
	app := fiber.New()
	limiter := NewRateLimiter(RateLimitConfig{
		Enabled:  true,
		Requests: 100, // global — generous
		Window:   time.Minute,
	})
	// Strict limit for the high-risk control mutations.
	limiter.SetRouteLimit("/v1/admin/delivery-control/pause", 2, time.Minute)
	app.Use(limiter.Handler())
	app.Post("/v1/admin/delivery-control/pause", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	// First two requests pass.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/delivery-control/pause", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, resp.StatusCode)
		}
	}

	// Third request is rate limited with Retry-After.
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/delivery-control/pause", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("third request: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}
	if ra := resp.Header.Get("Retry-After"); ra == "" {
		t.Fatal("expected Retry-After header on 429")
	}
}

// TestRateLimiter_DisabledWhenNotEnabled verifies the fail-open switch.
func TestRateLimiter_DisabledWhenNotEnabled(t *testing.T) {
	app := fiber.New()
	limiter := NewRateLimiter(RateLimitConfig{
		Enabled:  false,
		Requests: 1,
		Window:   time.Minute,
	})
	app.Use(limiter.Handler())
	app.Get("/v1/admin/delivery-control/status", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/delivery-control/status", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected 200 when disabled, got %d", i, resp.StatusCode)
		}
	}
}
