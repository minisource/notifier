package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/kavenegar/kavenegar-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestKavenegarClient builds a KavenegarClient whose HTTP calls are
// routed to the given httptest server instead of the real Kavenegar API.
func newTestKavenegarClient(t *testing.T, srv *httptest.Server, apiKey, template, sender string) *KavenegarClient {
	t.Helper()
	client := kavenegar.NewClient(apiKey)
	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)
	client.BaseURL = baseURL
	client.BaseClient = srv.Client()
	return &KavenegarClient{
		core:     kavenegar.NewWithClient(client),
		apiKey:   apiKey,
		template: template,
		sender:   sender,
	}
}

// TestKavenegarClient_PlainSendFallback guards the fix for providers
// configured with only an API key (no template): the client must fall back to
// the simple message API instead of returning "no template specified".
func TestKavenegarClient_PlainSendFallback(t *testing.T) {
	var gotEndpoint string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEndpoint = r.URL.Path
		_ = r.ParseForm()
		gotBody = r.PostFormValue("message")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"ok"},"entries":[{"messageid":1}]}`))
	}))
	defer srv.Close()

	k := newTestKavenegarClient(t, srv, "test-key", "", "") // no template, no sender
	err := k.SendMessage(map[string]string{"message": "Hello World"}, "+989011793041")
	require.NoError(t, err)

	assert.Equal(t, "/v1/test-key/sms/send.json", gotEndpoint)
	assert.Equal(t, "Hello World", gotBody)
}

// TestKavenegarClient_PlainSendFallback_UsesSender ensures the configured
// sender line is forwarded to the simple send API.
func TestKavenegarClient_PlainSendFallback_UsesSender(t *testing.T) {
	var gotSender string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotSender = r.PostFormValue("sender")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"ok"},"entries":[{"messageid":1}]}`))
	}))
	defer srv.Close()

	k := newTestKavenegarClient(t, srv, "test-key", "", "10004346")
	err := k.SendMessage(map[string]string{"message": "Hello"}, "+989011793041")
	require.NoError(t, err)
	assert.Equal(t, "10004346", gotSender)
}

// TestKavenegarClient_TemplateStillUsesLookup verifies the lookup path is
// unchanged when a template IS configured.
func TestKavenegarClient_TemplateStillUsesLookup(t *testing.T) {
	var gotEndpoint string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEndpoint = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"ok"},"entries":[{"messageid":1}]}`))
	}))
	defer srv.Close()

	k := newTestKavenegarClient(t, srv, "test-key", "verify", "")
	err := k.SendMessage(map[string]string{"token": "123456"}, "+989011793041")
	require.NoError(t, err)
	assert.True(t, strings.Contains(gotEndpoint, "/verify/lookup"), "expected lookup endpoint, got %s", gotEndpoint)
}

// TestKavenegarClient_PlainSend_EmptySenderOmitsParam guards the 412 fix:
// when no sender line is configured the "sender" form value must be omitted
// entirely (the SDK's Message.Send would send an empty sender= which Kavenegar
// rejects with APIError[412] "invalid sender").
func TestKavenegarClient_PlainSend_EmptySenderOmitsParam(t *testing.T) {
	var gotSenderSet bool
	var gotSender string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, gotSenderSet = r.PostForm["sender"]
		gotSender = r.PostFormValue("sender")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"ok"},"entries":[{"messageid":1}]}`))
	}))
	defer srv.Close()

	k := newTestKavenegarClient(t, srv, "test-key", "", "") // no template, no sender
	err := k.SendMessage(map[string]string{"message": "Hello"}, "+989011793041")
	require.NoError(t, err)

	assert.False(t, gotSenderSet, "sender param must be omitted when no sender line is configured")
	assert.Equal(t, "", gotSender)
}

// TestKavenegarClient_PlainSend_412SenderErrorMapping ensures a rejected
// sender line produces an actionable hint pointing at senderId.
func TestKavenegarClient_PlainSend_412SenderErrorMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"return":{"status":412,"message":"ارسال کننده نامعتبر است"}}`))
	}))
	defer srv.Close()

	// No sender configured: hint must say no senderId configured.
	k := newTestKavenegarClient(t, srv, "test-key", "", "")
	err := k.SendMessage(map[string]string{"message": "Hello"}, "+989011793041")
	require.Error(t, err)

	msg := err.Error()
	assert.Contains(t, msg, "APIError[412]")
	assert.Contains(t, msg, "no senderId configured")
	assert.Contains(t, msg, "senderId")

	// Sender configured but rejected: hint must name the configured line.
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"return":{"status":412,"message":"ارسال کننده نامعتبر است"}}`))
	}))
	defer srv2.Close()
	k2 := newTestKavenegarClient(t, srv2, "test-key", "", "30007777")
	err2 := k2.SendMessage(map[string]string{"message": "Hello"}, "+989011793041")
	require.Error(t, err2)

	msg2 := err2.Error()
	assert.Contains(t, msg2, "APIError[412]")
	assert.Contains(t, msg2, "sender line \"30007777\" rejected")
	assert.Contains(t, msg2, "not registered or approved")
}

// TestKavenegarClient_DescribeRequest_PlainSend verifies the real outbound
// request description for the plain-send path: POST to sms/send.json with the
// API key redacted, masked recipient, and the message preserved (no OTP).
func TestKavenegarClient_DescribeRequest_PlainSend(t *testing.T) {
	k := &KavenegarClient{apiKey: "secret-api-key", template: "", sender: "10004346"}
	method, reqURL, body := k.DescribeRequest(
		map[string]string{"message": "Hello World"},
		"+989011793041",
	)
	assert.Equal(t, "POST", method)
	assert.Equal(t, "https://api.kavenegar.com/v1/[REDACTED]/sms/send.json", reqURL)
	assert.NotContains(t, reqURL, "secret-api-key")
	assert.Contains(t, body, "sender=10004346")
	assert.Contains(t, body, "receptor=")
	assert.NotContains(t, body, "989011793041") // recipient masked
	assert.Contains(t, body, "message=Hello+World")
}

// TestKavenegarClient_DescribeRequest_PlainSendNoSender verifies the sender
// param is absent from the real request description when not configured.
func TestKavenegarClient_DescribeRequest_PlainSendNoSender(t *testing.T) {
	k := &KavenegarClient{apiKey: "secret-api-key", template: "", sender: ""}
	_, _, body := k.DescribeRequest(
		map[string]string{"body": "Hello"},
		"+989011793041",
	)
	assert.NotContains(t, body, "sender=")
	assert.Contains(t, body, "message=Hello")
}

// TestKavenegarClient_Check_NoSenderLineDegraded guards the degraded-state
// detection: a valid API key with no default sender line and no senderId/
// template must yield ErrNoSenderLine (health check reports degraded), because
// every plain send would be rejected with APIError[412].
func TestKavenegarClient_Check_NoSenderLineDegraded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/account/info.json"):
			_, _ = w.Write([]byte(`{"return":{"status":200,"message":"تایید شد"},"entries":{}}`))
		case strings.Contains(r.URL.Path, "/account/config.json"):
			// defaultsender empty + no entries sender config
			_, _ = w.Write([]byte(`{"return":{"status":200,"message":"تایید شد"},"entries":[{"defaultsender":""}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	// No template, no sender configured -> degraded.
	k := &KavenegarClient{apiKey: "test-key", baseURL: srv.URL}
	err := k.Check(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoSenderLine)

	// Template configured -> plain path not used, not degraded.
	k2 := &KavenegarClient{apiKey: "test-key", template: "verify", baseURL: srv.URL}
	require.NoError(t, k2.Check(context.Background()))

	// Sender configured -> plain sends would work, not degraded.
	k3 := &KavenegarClient{apiKey: "test-key", sender: "10004347", baseURL: srv.URL}
	require.NoError(t, k3.Check(context.Background()))
}

// TestKavenegarClient_Check_InvalidKey verifies the API key rejection path.
func TestKavenegarClient_Check_InvalidKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"return":{"status":403,"message":"درخواست غير مجاز است"},"entries":null}`))
	}))
	defer srv.Close()

	k := &KavenegarClient{apiKey: "bad-key", baseURL: srv.URL}
	err := k.Check(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rejected API key")
}

// TestKavenegarClient_DescribeRequest_PlainSend_OTPRedacts verifies that an
// OTP-shaped free-text message (all-digit, short) is redacted from the
// described request body even on the plain-send path.
func TestKavenegarClient_DescribeRequest_PlainSend_OTPRedacts(t *testing.T) {
	k := &KavenegarClient{apiKey: "secret-api-key", template: "", sender: ""}
	_, _, body := k.DescribeRequest(
		map[string]string{"message": "123456"},
		"+989011793041",
	)
	unescaped, err := url.QueryUnescape(body)
	require.NoError(t, err)
	assert.Contains(t, unescaped, "message=[REDACTED]")
	assert.NotContains(t, unescaped, "123456")
}

// TestKavenegarClient_DescribeRequest_Lookup verifies the lookup (OTP) request
// description redacts token values — only presence is reported. The form body
// is URL-encoded (the exact bytes sent to Kavenegar), so assertions unescape.
func TestKavenegarClient_DescribeRequest_Lookup(t *testing.T) {
	k := &KavenegarClient{apiKey: "secret-api-key", template: "verify", sender: ""}
	method, reqURL, body := k.DescribeRequest(
		map[string]string{"token": "123456", "token2": "45000"},
		"+989011793041",
	)
	assert.Equal(t, "POST", method)
	assert.Equal(t, "https://api.kavenegar.com/v1/[REDACTED]/verify/lookup.json", reqURL)

	unescaped, err := url.QueryUnescape(body)
	require.NoError(t, err)
	assert.Contains(t, unescaped, "template=verify")
	assert.Contains(t, unescaped, "token=[REDACTED]")
	assert.Contains(t, unescaped, "token2=[REDACTED]")
	assert.NotContains(t, unescaped, "123456")
	assert.NotContains(t, unescaped, "45000")
}

// TestKavenegarClient_PlainSend_413RecipientMapping ensures a rejected
// recipient (413) gets the actionable phone-format hint.
func TestKavenegarClient_PlainSend_413RecipientMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"return":{"status":413,"message":"گیرنده نامعتبر است"}}`))
	}))
	defer srv.Close()

	k := newTestKavenegarClient(t, srv, "test-key", "", "")
	err := k.SendMessage(map[string]string{"message": "Hello"}, "+989011793041")
	require.Error(t, err)

	msg := err.Error()
	assert.Contains(t, msg, "APIError[413]")
	assert.Contains(t, msg, "recipient number rejected")
}

// TestKavenegarClient_PlainSend_NoMessageText ensures a clear error when
// neither template nor message text is available.
func TestKavenegarClient_PlainSend_NoMessageText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	k := newTestKavenegarClient(t, srv, "test-key", "", "")
	err := k.SendMessage(map[string]string{}, "+989011793041")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no message text")
}

// TestKavenegarClient_Lookup_431StructureMapping ensures a Kavenegar
// APIError[431] ("code structure is not correct") is mapped to an actionable
// hint pointing at the template token format instead of the raw misleading
// provider message.
func TestKavenegarClient_Lookup_431StructureMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"return":{"status":431,"message":"ساختار کد صحیح نمی باشد"}}`))
	}))
	defer srv.Close()

	k := newTestKavenegarClient(t, srv, "test-key", "verify", "")
	err := k.SendMessage(map[string]string{"token": "4321"}, "+989011793041")
	require.Error(t, err)

	msg := err.Error()
	assert.Contains(t, msg, "APIError[431]")
	assert.Contains(t, msg, "code/token structure rejected")
	assert.Contains(t, msg, "template's expected format")
}

// TestKavenegarClient_Lookup_UnresolvedPlaceholderGuard guards the fix for
// template-based sends where the variables never reached the provider params:
// the token still contains {{...}} markers (e.g. the raw body
// "Your verification code is: {{code}}"). Instead of sending it to Kavenegar
// (which replies with a misleading APIError[431]), the client must fail fast
// with a clear message.
func TestKavenegarClient_Lookup_UnresolvedPlaceholderGuard(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"ok"}}`))
	}))
	defer srv.Close()

	k := newTestKavenegarClient(t, srv, "test-key", "verify", "")
	err := k.SendMessage(map[string]string{"token": "Your verification code is: {{code}}"}, "+989011793041")
	require.Error(t, err)
	assert.False(t, hit, "no provider request must be sent when the token is unresolved")
	assert.Contains(t, err.Error(), "template variables not resolved")
	assert.Contains(t, err.Error(), "{{code}}")
}
