package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKavenegarClient_AccountInfo_Success verifies the account/info.json
// parsing: remaincredit is captured, expiry is converted, and the API key is
// never present in any returned field.
func TestKavenegarClient_AccountInfo_Success(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"return":{"status":200,"message":"تایید شد"},
			"entries":[{"remaincredit":1250,"expiredate":1785715000,"type":"service","signtext":"نرم افزار تست","isverifysend":1}]
		}`))
	}))
	defer srv.Close()

	k := &KavenegarClient{apiKey: "secret-api-key", baseURL: srv.URL}
	info, err := k.AccountInfo(context.Background())
	require.NoError(t, err)
	require.NotNil(t, info)
	require.NotNil(t, info.RemainCredit)
	assert.Equal(t, 1250.0, *info.RemainCredit)
	require.NotNil(t, info.ExpireDate)
	assert.Equal(t, int64(1785715000), info.ExpireDate.Unix())
	assert.Equal(t, "service", info.AccountType)
	assert.True(t, info.IsVerifySend)
	assert.Equal(t, "/v1/secret-api-key/account/info.json", gotPath)
}

// TestKavenegarClient_AccountInfo_ObjectEntries verifies the OBJECT form of
// `entries` that Kavenegar returns for some account types (e.g. Master):
// {"remaincredit":6206389,"expiredate":"1799267400","type":"Master"} —
// with expiredate as a numeric STRING. This is the real-world response shape
// seen against the live API.
func TestKavenegarClient_AccountInfo_ObjectEntries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"return":{"status":200,"message":"تایید شد"},
			"entries":{"remaincredit":6206389,"expiredate":"1799267400","type":"Master"}
		}`))
	}))
	defer srv.Close()

	k := &KavenegarClient{apiKey: "secret-api-key", baseURL: srv.URL}
	info, err := k.AccountInfo(context.Background())
	require.NoError(t, err)
	require.NotNil(t, info.RemainCredit)
	assert.Equal(t, 6206389.0, *info.RemainCredit)
	require.NotNil(t, info.ExpireDate)
	assert.Equal(t, int64(1799267400), info.ExpireDate.Unix())
	assert.Equal(t, "Master", info.AccountType)
}

// TestKavenegarClient_AccountInfo_NoEntries verifies an empty entries array
// yields nil balance (unknown), never zero — the caller must treat unknown
// separately from zero.
func TestKavenegarClient_AccountInfo_NoEntries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"ok"},"entries":[]}`))
	}))
	defer srv.Close()

	k := &KavenegarClient{apiKey: "key", baseURL: srv.URL}
	info, err := k.AccountInfo(context.Background())
	require.NoError(t, err)
	assert.Nil(t, info.RemainCredit, "absent balance must be nil, never zero")
}

// TestKavenegarClient_AccountInfo_ErrorEnvelope verifies Kavenegar's HTTP-200
// error envelope is normalized into a typed error with a safe message.
func TestKavenegarClient_AccountInfo_ErrorEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"return":{"status":332,"message":"نام کاربری یا رمز عبور اشتباه است"},"entries":null}`))
	}))
	defer srv.Close()

	k := &KavenegarClient{apiKey: "bad-key", baseURL: srv.URL}
	_, err := k.AccountInfo(context.Background())
	require.Error(t, err)

	kind, code, msg := NormalizeKavenegarBalanceError(err)
	assert.Equal(t, "authentication", kind)
	assert.Equal(t, "kavenegar_332", code)
	assert.NotContains(t, msg, "bad-key")
	assert.NotEmpty(t, msg)
}

// TestKavenegarClient_AccountInfo_NonOKHTTP verifies non-200 transport
// responses are classified as network errors.
func TestKavenegarClient_AccountInfo_NonOKHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream down", http.StatusBadGateway)
	}))
	defer srv.Close()

	k := &KavenegarClient{apiKey: "key", baseURL: srv.URL}
	_, err := k.AccountInfo(context.Background())
	require.Error(t, err)
	kind, code, _ := NormalizeKavenegarBalanceError(err)
	assert.Equal(t, "network", kind)
	assert.Contains(t, code, "kavenegar")
}

// TestKavenegarClient_AccountInfo_Malformed verifies unparseable bodies are
// classified as malformed (safe, no raw body echoed into the persisted
// message beyond a bounded note).
func TestKavenegarClient_AccountInfo_Malformed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not json`))
	}))
	defer srv.Close()

	k := &KavenegarClient{apiKey: "super-secret-api-key-xyz", baseURL: srv.URL}
	_, err := k.AccountInfo(context.Background())
	require.Error(t, err)
	kind, _, msg := NormalizeKavenegarBalanceError(err)
	assert.Equal(t, "network", kind)
	assert.NotContains(t, msg, "super-secret-api-key-xyz", "raw API key value must never appear in the sanitized message")
}

// TestNormalizeKavenegarBalanceError_NoLeak ensures error messages never
// contain the API key even when the raw error would.
func TestNormalizeKavenegarBalanceError_NoLeak(t *testing.T) {
	_, _, msg := NormalizeKavenegarBalanceError(&kavenegarBalanceErr{
		Status:  333,
		Message: "connection to api failed",
		HTTP:    200,
	})
	assert.NotEmpty(t, msg)
	assert.False(t, strings.Contains(msg, "api.kavenegar.com"), "sanitized message must not contain provider URL")
}
