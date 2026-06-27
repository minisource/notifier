//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/minisource/go-common/testing/e2e"
)

func TestNotifier_TemplatesCRUD(t *testing.T) {
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", defaultAuthURL)
	notifierURL := e2e.BaseURLFromEnv("NOTIFIER_BASE_URL", defaultNotifierURL)

	userToken := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	userH := e2e.Bearer(userToken)

	notifier := e2e.NewClient(notifierURL, userH)
	notifier.RequireUp(t, "/api/v1/health/")

	name := fmt.Sprintf("e2e-template-%d", time.Now().UnixNano())
	resp, body, err := notifier.Do(http.MethodPost, "/api/v1/templates", map[string]any{
		"name": name, "type": "in_app", "subject": "e2e", "body": "hello {{name}}",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	templateID := e2e.ExtractID(t, body)

	resp, body, err = notifier.Do(http.MethodGet, "/api/v1/templates/"+templateID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = notifier.Do(http.MethodGet, "/api/v1/templates", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = notifier.Do(http.MethodPut, "/api/v1/templates/"+templateID, map[string]any{
		"name": name, "type": "in_app", "subject": "e2e updated", "body": "updated body",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = notifier.Do(http.MethodDelete, "/api/v1/templates/"+templateID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
}
