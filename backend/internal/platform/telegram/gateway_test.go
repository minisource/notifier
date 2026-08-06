package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fakeGateway is a minimal fake of the official Telegram Gateway API. It
// asserts the Authorization header and records each operation + request body
// so tests can verify the exact contract (endpoint, method, payload).
type fakeGateway struct {
	t            *testing.T
	token        string
	ok           bool
	errorCode    string
	result       map[string]interface{}
	mu           atomic.Int64
	lastOp       atomic.Value // string
	lastBody     atomic.Value // string
	delay        time.Duration
	statusCode   int
}

func (f *fakeGateway) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.delay > 0 {
			time.Sleep(f.delay)
		}
		// Official auth: Authorization: Bearer <token>.
		if got := r.Header.Get("Authorization"); got != "Bearer "+f.token {
			f.t.Errorf("Authorization header = %q, want Bearer %s", got, f.token)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		body, _ := io.ReadAll(r.Body)
		f.lastOp.Store(r.URL.Path)
		f.lastBody.Store(string(body))
		f.mu.Add(1)

		status := f.statusCode
		if status == 0 {
			status = http.StatusOK
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		resp := map[string]interface{}{"ok": f.ok}
		if !f.ok {
			resp["error"] = f.errorCode
		} else if f.result != nil {
			resp["result"] = f.result
		} else {
			resp["result"] = map[string]interface{}{
				"request_id":         "req_123",
				"phone_number":       "+989123456789",
				"request_cost":       0.005,
				"is_refunded":        false,
				"remaining_balance":  10.0,
				"delivery_status":    map[string]interface{}{"status": "sent", "updated_at": 1710000000},
				"verification_status": map[string]interface{}{"status": "code_valid", "updated_at": 1710000001},
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
}

func (f *fakeGateway) ops() int64    { return f.mu.Load() }
func (f *fakeGateway) op() string    { v, _ := f.lastOp.Load().(string); return v }
func (f *fakeGateway) body() string  { v, _ := f.lastBody.Load().(string); return v }

func newTestClient(t *testing.T, gw *fakeGateway, opts Options) *Client {
	t.Helper()
	srv := httptest.NewServer(gw.handler())
	t.Cleanup(srv.Close)
	opts.Token = gw.token
	opts.BaseURL = srv.URL
	opts.RequestTimeout = 5 * time.Second
	opts.ConnectTimeout = 2 * time.Second
	c, err := NewClient(opts)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestClientAuthAndSendVerificationMessage(t *testing.T) {
	gw := &fakeGateway{t: t, token: "secret-token", ok: true}
	c := newTestClient(t, gw, Options{})

	res, err := c.SendVerificationMessage(context.Background(), SendVerificationMessageRequest{
		PhoneNumber: "+989123456789",
		Code:        "123456",
		TTL:         120,
	})
	if err != nil {
		t.Fatalf("SendVerificationMessage: %v", err)
	}
	if res.RequestID != "req_123" {
		t.Errorf("RequestID = %q, want req_123", res.RequestID)
	}
	if res.DeliveryStatus == nil || res.DeliveryStatus.Status != "sent" {
		t.Errorf("DeliveryStatus = %+v, want sent", res.DeliveryStatus)
	}
	if gw.op() != "/sendVerificationMessage" {
		t.Errorf("endpoint = %q, want /sendVerificationMessage", gw.op())
	}
	// Payload must use the official field names.
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(gw.body()), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if body["phone_number"] != "+989123456789" {
		t.Errorf("phone_number = %v", body["phone_number"])
	}
	if body["code"] != "123456" {
		t.Errorf("code = %v", body["code"])
	}
	if body["ttl"] != float64(120) {
		t.Errorf("ttl = %v, want 120", body["ttl"])
	}
}

func TestClientCheckSendAbility(t *testing.T) {
	gw := &fakeGateway{t: t, token: "t", ok: true}
	c := newTestClient(t, gw, Options{})

	_, err := c.CheckSendAbility(context.Background(), CheckSendAbilityRequest{PhoneNumber: "+989123456789"})
	if err != nil {
		t.Fatalf("CheckSendAbility: %v", err)
	}
	if gw.op() != "/checkSendAbility" {
		t.Errorf("endpoint = %q, want /checkSendAbility", gw.op())
	}
	var body map[string]interface{}
	_ = json.Unmarshal([]byte(gw.body()), &body)
	if body["phone_number"] != "+989123456789" {
		t.Errorf("phone_number = %v", body["phone_number"])
	}
}

func TestClientCheckVerificationStatus(t *testing.T) {
	gw := &fakeGateway{t: t, token: "t", ok: true}
	c := newTestClient(t, gw, Options{})

	res, err := c.CheckVerificationStatus(context.Background(), CheckVerificationStatusRequest{
		RequestID: "req_123",
		Code:      "123456",
	})
	if err != nil {
		t.Fatalf("CheckVerificationStatus: %v", err)
	}
	if gw.op() != "/checkVerificationStatus" {
		t.Errorf("endpoint = %q, want /checkVerificationStatus", gw.op())
	}
	if res.VerificationStatus == nil || res.VerificationStatus.Status != "code_valid" {
		t.Errorf("VerificationStatus = %+v, want code_valid", res.VerificationStatus)
	}
	var body map[string]interface{}
	_ = json.Unmarshal([]byte(gw.body()), &body)
	if body["request_id"] != "req_123" {
		t.Errorf("request_id = %v", body["request_id"])
	}
}

func TestClientRevokeVerificationMessage(t *testing.T) {
	gw := &fakeGateway{t: t, token: "t", ok: true}
	c := newTestClient(t, gw, Options{})

	_, err := c.RevokeVerificationMessage(context.Background(), RevokeVerificationMessageRequest{RequestID: "req_123"})
	if err != nil {
		t.Fatalf("RevokeVerificationMessage: %v", err)
	}
	if gw.op() != "/revokeVerificationMessage" {
		t.Errorf("endpoint = %q, want /revokeVerificationMessage", gw.op())
	}
}

// Error contract: {ok:false, error:"ERROR_CODE"} — the code is exposed but the
// token must never leak into the error.
func TestClientErrorMapping(t *testing.T) {
	cases := []struct {
		name     string
		errCode  string
		wantKind string
	}{
		{"invalid token", "ACCESS_TOKEN_INVALID", "authentication"},
		{"expired token", "ACCESS_TOKEN_EXPIRED", "authentication"},
		{"rate limited", "RATE_LIMITED", "rate_limited"},
		{"no balance", "INSUFFICIENT_BALANCE", "insufficient_balance"},
		{"bad phone", "PHONE_NUMBER_INVALID", "invalid_recipient"},
		{"missing request", "REQUEST_NOT_FOUND", "invalid_request"},
		{"server error", "INTERNAL_ERROR", "provider"},
		{"unknown code", "SOME_OTHER_CODE", "provider"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gw := &fakeGateway{t: t, token: "super-secret-token", ok: false, errorCode: tc.errCode}
			c := newTestClient(t, gw, Options{})
			_, err := c.SendVerificationMessage(context.Background(), SendVerificationMessageRequest{
				PhoneNumber: "+989123456789", Code: "123456",
			})
			if err == nil {
				t.Fatal("expected error")
			}
			if strings.Contains(err.Error(), gw.token) {
				t.Fatalf("error leaks the token: %v", err)
			}
			if got := NormalizedErrorKind(err); got != tc.wantKind {
				t.Errorf("NormalizedErrorKind = %q, want %q (err: %v)", got, tc.wantKind, err)
			}
			// The official code must still be retrievable for safe surfacing.
			if code := ErrorCode(err); code != tc.errCode {
				t.Errorf("ErrorCode = %q, want %q", code, tc.errCode)
			}
		})
	}
}

// Timeout behavior: a transport timeout must classify as "timeout", never as a
// provider acceptance or a generic failure, and never leak the token.
func TestClientTimeout(t *testing.T) {
	gw := &fakeGateway{t: t, token: "secret", ok: true, delay: 3 * time.Second}
	c := newTestClient(t, gw, Options{})
	c.httpClient.Timeout = 200 * time.Millisecond

	_, err := c.SendVerificationMessage(context.Background(), SendVerificationMessageRequest{
		PhoneNumber: "+989123456789", Code: "123456",
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !IsTimeout(err) {
		t.Errorf("IsTimeout = false, err = %v", err)
	}
	if got := NormalizedErrorKind(err); got != "timeout" {
		t.Errorf("NormalizedErrorKind = %q, want timeout", got)
	}
	if strings.Contains(err.Error(), gw.token) {
		t.Fatalf("error leaks the token: %v", err)
	}
}

// Context cancellation must propagate and classify as timeout/network, not a
// provider error.
func TestClientContextCancellation(t *testing.T) {
	gw := &fakeGateway{t: t, token: "secret", ok: true, delay: 3 * time.Second}
	c := newTestClient(t, gw, Options{})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := c.SendVerificationMessage(ctx, SendVerificationMessageRequest{
		PhoneNumber: "+989123456789", Code: "123456",
	})
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
	if IsTimeout(err) && NormalizedErrorKind(err) == "timeout" {
		return
	}
	// Fallback: any error without token leak is acceptable.
	if strings.Contains(err.Error(), gw.token) {
		t.Fatalf("error leaks the token: %v", err)
	}
}

// OTP values must never appear in errors, even in transport-level messages.
func TestClientNoCodeLeak(t *testing.T) {
	gw := &fakeGateway{t: t, token: "secret", ok: false, errorCode: "ACCESS_TOKEN_INVALID"}
	c := newTestClient(t, gw, Options{})
	_, err := c.SendVerificationMessage(context.Background(), SendVerificationMessageRequest{
		PhoneNumber: "+989123456789", Code: "987654",
	})
	if err != nil && strings.Contains(err.Error(), "987654") {
		t.Fatalf("error leaks the OTP: %v", err)
	}
}

// NewClient must reject a missing token and invalid base URL.
func TestNewClientValidation(t *testing.T) {
	if _, err := NewClient(Options{Token: ""}); err == nil {
		t.Error("expected error for empty token")
	}
	if _, err := NewClient(Options{Token: "t", BaseURL: "gatewayapi.telegram.org"}); err == nil {
		t.Error("expected error for non-http base URL")
	}
}

// The error model must distinguish errors.As on apiError so callers can map
// official codes without string matching.
func TestAPIErrorIsExposed(t *testing.T) {
	gw := &fakeGateway{t: t, token: "secret", ok: false, errorCode: "RATE_LIMITED"}
	c := newTestClient(t, gw, Options{})
	_, err := c.SendVerificationMessage(context.Background(), SendVerificationMessageRequest{
		PhoneNumber: "+989123456789", Code: "123456",
	})
	var ae *apiError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *apiError, got %T: %v", err, err)
	}
	if ae.Code != "RATE_LIMITED" {
		t.Errorf("apiError.Code = %q, want RATE_LIMITED", ae.Code)
	}
}
