package sms_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFakeKavenegarLookupAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/verify/lookup.json") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"return": map[string]interface{}{
				"status": 200,
				"entries": []map[string]interface{}{
					{"messageid": 1, "status": 5, "statustext": "Sent"},
				},
			},
		})
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/v1/test-key/verify/lookup.json?receptor=09123456789&token=654321&template=verify")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"status":200`) {
		t.Fatalf("unexpected body: %s", body)
	}
}
