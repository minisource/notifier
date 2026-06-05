//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/minisource/go-common/testing/e2e"
)

func TestNotifier_CreateMarkReadAndPreferences(t *testing.T) {
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", defaultAuthURL)
	notifierURL := e2e.BaseURLFromEnv("NOTIFIER_BASE_URL", defaultNotifierURL)

	auth := e2e.NewClient(authURL, nil)
	auth.RequireUp(t, "/health")
	svcToken := e2e.ServiceToken(t, authURL, "auth-service", "auth-service-secret-key")
	svcH := e2e.Bearer(svcToken)

	userToken := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	adminID := fetchAdminID(t, auth, userToken)

	notifier := e2e.NewClient(notifierURL, svcH)
	notifier.RequireUp(t, "/api/v1/health/")

	resp, body, err := notifier.Do(http.MethodPost, "/api/v1/service/notifications", map[string]any{
		"userId": adminID, "type": "in_app", "body": "mark-read e2e", "subject": "e2e", "priority": "normal",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	notifID := e2e.ExtractID(t, body)

	resp, body, err = notifier.Do(http.MethodPut, "/api/v1/service/notifications/"+notifID+"/read", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = notifier.Do(http.MethodGet, "/api/v1/service/notifications/user/"+adminID+"/unread", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	userH := e2e.Bearer(userToken)
	resp, body, err = notifier.WithHeaders(userH).Do(http.MethodPut, "/api/v1/preferences/user/"+adminID, map[string]any{
		"type": "in_app", "isEnabled": true, "allowInstant": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)

	resp, body, err = notifier.WithHeaders(userH).Do(http.MethodGet, "/api/v1/preferences/user/"+adminID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)
}

func fetchAdminID(t *testing.T, c *e2e.Client, token string) string {
	t.Helper()
	resp, body, err := c.WithHeaders(e2e.Bearer(token)).Do(http.MethodGet, "/api/v1/users/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)
	var parsed map[string]any
	e2e.ParseJSON(t, body, &parsed)
	id := e2e.GetString(parsed, "data", "id")
	if id == "" {
		id = e2e.GetString(parsed, "id")
	}
	if id == "" {
		t.Fatalf("no user id: %s", string(body))
	}
	return id
}
