//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/minisource/go-common/testing/e2e"
)

const (
	defaultNotifierURL = "http://127.0.0.1:9002"
	defaultAuthURL     = "http://127.0.0.1:9001"
)

func TestNotifier_API(t *testing.T) {
	c := e2e.NewClient(e2e.BaseURLFromEnv("NOTIFIER_BASE_URL", defaultNotifierURL), nil)
	c.RequireUp(t, "/api/v1/health/")

	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", defaultAuthURL)
	svcToken := e2e.ServiceToken(t, authURL, "auth-service", "auth-service-secret-key")
	svcH := e2e.Bearer(svcToken)
	userToken := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	userH := e2e.Bearer(userToken)

	userID := "742fe13c-b5f2-430f-a7e4-9a096d2f5669"

	c.RunCases(t, []e2e.Case{
		{Name: "health", Method: http.MethodGet, Path: "/api/v1/health/", WantCode: []int{http.StatusOK}},
		{Name: "service_create_notification", Method: http.MethodPost, Path: "/api/v1/service/notifications", Headers: svcH, Body: map[string]any{
			"userId": userID, "type": "in_app", "body": "e2e test", "priority": "normal",
		}, WantCode: []int{http.StatusOK, http.StatusCreated}},
		{Name: "service_list_user", Method: http.MethodGet, Path: "/api/v1/service/notifications/user/" + userID, Headers: svcH, WantCode: []int{http.StatusOK}},
		{Name: "service_unread", Method: http.MethodGet, Path: "/api/v1/service/notifications/user/" + userID + "/unread", Headers: svcH, WantCode: []int{http.StatusOK}},
		{Name: "templates_list", Method: http.MethodGet, Path: "/api/v1/templates", Headers: userH, WantCode: []int{http.StatusOK, http.StatusUnauthorized, http.StatusForbidden}},
		{Name: "preferences_get", Method: http.MethodGet, Path: "/api/v1/preferences/user/" + userID, Headers: svcH, WantCode: []int{http.StatusOK, http.StatusNotFound, http.StatusUnauthorized}},
		{Name: "notifications_user", Method: http.MethodGet, Path: "/api/v1/notifications/user/" + userID, Headers: svcH, WantCode: []int{http.StatusOK, http.StatusUnauthorized, http.StatusForbidden}},
	})
}

func TestNotifier_MockSMSProvider(t *testing.T) {
	// Unit-level mock SMS is in internal/platform/sms; run: go test ./internal/platform/sms/ -run Mock
	t.Log("mock SMS provider: go test ./internal/platform/sms/ -run TestMockSMSProvider")
}
