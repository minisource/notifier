package providers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// testTelegramConfig returns an adapter config pointed at a fake server with a
// bounded timeout, so SendMessage can be exercised end-to-end.
func testTelegramConfig(baseURL string) TelegramGatewayClientConfig {
	return TelegramGatewayClientConfig{
		BaseURL:          baseURL,
		RequestTimeout:   5 * time.Second,
		ConnectTimeout:   2 * time.Second,
		MaxResponseBytes: 1 << 20,
		TTL:              120,
		CheckPhone:       "+989123456789",
	}
}

// DescribeRequest must redact the OTP code and mask the phone while keeping the
// official endpoint + TTL visible. The token is never part of the URL/body.
func TestTelegramGatewayDescribeRequestRedaction(t *testing.T) {
	client, err := GetTelegramGatewayClient("secret-token", testTelegramConfig("https://gatewayapi.telegram.org"))
	if err != nil {
		t.Fatalf("GetTelegramGatewayClient: %v", err)
	}

	method, url, body := client.DescribeRequest(
		map[string]string{"code": "123456", "ttl": "120"},
		"+989123456789",
	)
	if method != "POST" {
		t.Errorf("method = %q, want POST", method)
	}
	if !strings.HasSuffix(url, "/sendVerificationMessage") {
		t.Errorf("url = %q, want suffix /sendVerificationMessage", url)
	}
	if strings.Contains(body, "123456") {
		t.Errorf("body leaks the OTP: %s", body)
	}
	if !strings.Contains(body, "[REDACTED]") {
		t.Errorf("body does not contain the redaction marker: %s", body)
	}
	if strings.Contains(body, "+989123456789") || strings.Contains(body, "9123456789") {
		t.Errorf("body leaks the full phone number: %s", body)
	}
	if !strings.Contains(body, "ttl") {
		t.Errorf("body should keep ttl for operators: %s", body)
	}
	if strings.Contains(body, "secret-token") || strings.Contains(url, "secret-token") {
		t.Errorf("request leaks the API token: url=%q body=%s", url, body)
	}
}

// The adapter must reject sends without a code (MiniSource owns the OTP, a
// send without a code would silently misuse code_length).
func TestTelegramGatewaySendRequiresCode(t *testing.T) {
	client, err := GetTelegramGatewayClient("secret-token", testTelegramConfig("https://gatewayapi.telegram.org"))
	if err != nil {
		t.Fatalf("GetTelegramGatewayClient: %v", err)
	}
	err = client.SendMessage(map[string]string{}, "+989123456789")
	if err == nil {
		t.Fatal("expected error when no OTP code is provided")
	}
	if !strings.Contains(err.Error(), "OTP code is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// The adapter must reject sends with no recipient, mirroring other providers.
func TestTelegramGatewaySendNoRecipient(t *testing.T) {
	client, err := GetTelegramGatewayClient("secret-token", testTelegramConfig("https://gatewayapi.telegram.org"))
	if err != nil {
		t.Fatalf("GetTelegramGatewayClient: %v", err)
	}
	err = client.SendMessage(map[string]string{"code": "123456"})
	if err == nil {
		t.Fatal("expected error when no recipient is provided")
	}
}

// End-to-end: SendMessage must hit /sendVerificationMessage on the configured
// base URL with the official payload (phone_number, code, ttl) and Bearer auth,
// and mask the phone in surfaced errors.
func TestTelegramGatewaySendMessageFakeServer(t *testing.T) {
	var mu atomic.Int64
	var lastPath, lastAuth, lastBody atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Add(1)
		lastPath.Store(r.URL.Path)
		lastAuth.Store(r.Header.Get("Authorization"))
		body, _ := io.ReadAll(r.Body)
		lastBody.Store(string(body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{"request_id":"req_abc","request_cost":0.005,"remaining_balance":9.5,"delivery_status":{"status":"sent","updated_at":1710000000}}}`))
	}))
	defer srv.Close()

	client, err := GetTelegramGatewayClient("tok-123", testTelegramConfig(srv.URL))
	if err != nil {
		t.Fatalf("GetTelegramGatewayClient: %v", err)
	}

	// Success path.
	if err := client.SendMessage(map[string]string{"code": "424242", "ttl": "180"}, "+989123456789"); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if mu.Load() != 1 {
		t.Fatalf("fake server received %d requests, want 1", mu.Load())
	}
	if p, _ := lastPath.Load().(string); p != "/sendVerificationMessage" {
		t.Errorf("path = %q, want /sendVerificationMessage", p)
	}
	if a, _ := lastAuth.Load().(string); a != "Bearer tok-123" {
		t.Errorf("Authorization = %q, want Bearer tok-123", a)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(lastBody.Load().(string)), &payload); err != nil {
		t.Fatalf("request body not JSON: %v", err)
	}
	if payload["phone_number"] != "+989123456789" {
		t.Errorf("phone_number = %v", payload["phone_number"])
	}
	if payload["code"] != "424242" {
		t.Errorf("code = %v", payload["code"])
	}
	if payload["ttl"] != float64(180) {
		t.Errorf("ttl = %v, want 180 (param override)", payload["ttl"])
	}

	// Error path: provider rejection must not leak the token; the phone is
	// masked in the surfaced message.
	mu.Store(0)
	handler := srv.Config.Handler
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":false,"error":"ACCESS_TOKEN_INVALID"}`))
	})
	defer func() { srv.Config.Handler = handler }()
	err = client.SendMessage(map[string]string{"code": "424242"}, "+989123456789")
	if err == nil {
		t.Fatal("expected error on provider rejection")
	}
	if strings.Contains(err.Error(), "tok-123") {
		t.Errorf("error leaks the token: %v", err)
	}
	if strings.Contains(err.Error(), "+989123456789") || strings.Contains(err.Error(), "9123456789") {
		t.Errorf("error leaks the full phone: %v", err)
	}
}
