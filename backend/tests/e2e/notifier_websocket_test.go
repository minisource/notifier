//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/minisource/go-common/testing/e2e"
)

func TestNotifier_WebSocketRequiresToken(t *testing.T) {
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	token := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	c := e2e.NewClient(e2e.BaseURLFromEnv("NOTIFIER_BASE_URL", "http://127.0.0.1:9002"), nil)
	c.RequireUp(t, "/api/v1/health/")

	resp, body, err := c.Do(http.MethodGet, "/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusUnauthorized, http.StatusBadRequest, http.StatusUpgradeRequired, http.StatusNotFound)

	resp, body, err = c.Do(http.MethodGet, "/ws?token="+token, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusUnauthorized, http.StatusBadRequest, http.StatusUpgradeRequired, http.StatusOK)
}
