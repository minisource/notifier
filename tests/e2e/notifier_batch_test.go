//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/minisource/go-common/testing/e2e"
)

func TestNotifier_BatchNotifications(t *testing.T) {
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

	resp, body, err := notifier.Do(http.MethodPost, "/api/v1/service/notifications/batch", map[string]any{
		"notifications": []map[string]any{
			{"userId": adminID, "type": "in_app", "body": "batch-1", "subject": "e2e", "priority": "normal"},
			{"userId": adminID, "type": "in_app", "body": "batch-2", "subject": "e2e", "priority": "normal"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusNotFound)

	// User-facing batch route (if exposed)
	userNotifier := e2e.NewClient(notifierURL, e2e.Bearer(userToken))
	resp, body, err = userNotifier.Do(http.MethodPost, "/api/v1/notifications/batch", map[string]any{
		"notifications": []map[string]any{
			{"userId": adminID, "type": "in_app", "body": "user-batch", "subject": "e2e", "priority": "normal"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusUnauthorized, http.StatusForbidden)
}
