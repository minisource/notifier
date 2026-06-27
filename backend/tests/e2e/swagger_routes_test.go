//go:build e2e

package e2e_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/minisource/go-common/testing/e2e"
)

func TestNotifier_SwaggerRouteSmoke(t *testing.T) {
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", defaultAuthURL)
	token := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	c := e2e.NewClient(e2e.BaseURLFromEnv("NOTIFIER_BASE_URL", "http://127.0.0.1:9002"), e2e.Bearer(token))
	c.RequireUp(t, "/api/v1/health/")
	_, file, _, _ := runtime.Caller(0)
	doc := filepath.Join(filepath.Dir(file), "..", "..", "docs", "swagger.json")
	e2e.RunSwaggerSmoke(t, c, doc, e2e.Bearer(token), false)
}
