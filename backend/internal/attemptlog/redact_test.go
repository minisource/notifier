package attemptlog

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizeHeaderMap_RedactsSecrets(t *testing.T) {
	in := map[string]string{
		"Authorization": "Bearer secret-token",
		"Cookie":        "session=abc",
		"X-Request-Id":  "req-123",
		"Content-Type":  "application/json",
	}
	out := SanitizeHeaderMap(in)
	if out["Authorization"] != MarkerRedacted {
		t.Errorf("Authorization should be redacted, got %q", out["Authorization"])
	}
	if out["Cookie"] != MarkerRedacted {
		t.Errorf("Cookie should be redacted, got %q", out["Cookie"])
	}
	if out["X-Request-Id"] != "req-123" {
		t.Errorf("X-Request-Id should be preserved, got %q", out["X-Request-Id"])
	}
	if out["Content-Type"] != "application/json" {
		t.Errorf("Content-Type should be preserved, got %q", out["Content-Type"])
	}
}

func TestSanitizeQueryString_RedactsSecrets(t *testing.T) {
	out := SanitizeQueryString("api_key=abc123&phone=09123456789&template=verify")
	if strings.Contains(out, "abc123") {
		t.Errorf("api_key value leaked: %q", out)
	}
	if !strings.Contains(out, MarkerRedacted) {
		t.Errorf("expected redaction marker, got %q", out)
	}
	if !strings.Contains(out, "09123456789") {
		t.Errorf("non-sensitive query value was redacted: %q", out)
	}
	if !strings.Contains(out, "verify") {
		t.Errorf("non-sensitive query value was lost: %q", out)
	}
}

func TestSanitizeURL_RedactsQuerySecrets(t *testing.T) {
	out := SanitizeURL("https://api.kavenegar.com/v1/ABC123/verify/lookup.json?receptor=09&token=654321&template=verify")
	// Query-string secrets must be redacted; safe query values preserved.
	if strings.Contains(out, "654321") {
		t.Errorf("token value leaked: %q", out)
	}
	if !strings.Contains(out, "token="+MarkerRedacted) {
		t.Errorf("token param should carry redaction marker: %q", out)
	}
	if !strings.Contains(out, "receptor=09") {
		t.Errorf("non-sensitive query value lost: %q", out)
	}
	if !strings.Contains(out, "template=verify") {
		t.Errorf("template value lost: %q", out)
	}
}

func TestSanitizeURL_RedactsPathCredential(t *testing.T) {
	// Kavenegar-style URL embeds the API key in the path.
	out := SanitizeURL("https://api.kavenegar.com/v1/ABC123def456/verify/lookup.json")
	if strings.Contains(out, "ABC123def456") {
		t.Errorf("path API key leaked: %q", out)
	}
	if !strings.Contains(out, "verify") || !strings.Contains(out, "lookup.json") {
		t.Errorf("safe path segments lost: %q", out)
	}
}

func TestSanitizeURL_PreservesShortPathSegments(t *testing.T) {
	out := SanitizeURL("https://api.example.com/v1/verify/lookup.json")
	if strings.Contains(out, MarkerRedacted) {
		t.Errorf("short safe segments should not be redacted: %q", out)
	}
}

func TestSanitizeBody_RedactsAndTruncates(t *testing.T) {
	body := `{"apiKey":"secret-key","code":"123456","template":"verify","data":{"name":"Ali"}}`
	opts := RedactionOptions{MaxBodyBytes: 4096, MaxPreviewChars: 200}
	sanitized, truncated, orig, captured := SanitizeBody(body, opts)

	if strings.Contains(sanitized, "secret-key") {
		t.Errorf("apiKey leaked in sanitized body: %q", sanitized)
	}
	if strings.Contains(sanitized, "123456") {
		t.Errorf("OTP/code leaked in sanitized body: %q", sanitized)
	}
	if !strings.Contains(sanitized, "Ali") {
		t.Errorf("non-sensitive data lost: %q", sanitized)
	}
	if truncated {
		t.Errorf("body should not be truncated at 4096 limit")
	}
	if orig != len(body) {
		t.Errorf("original size mismatch: %d != %d", orig, len(body))
	}
	if captured == 0 {
		t.Errorf("captured size should be > 0")
	}

	// Truncation
	small := RedactionOptions{MaxBodyBytes: 20}
	truncBody, truncFlag, _, captured2 := SanitizeBody(body, small)
	if !truncFlag {
		t.Errorf("expected truncation flag with 20-byte limit")
	}
	suffix := "\n...[TRUNCATED]"
	if len(truncBody) > 20+len(suffix) {
		t.Errorf("truncated body too long: %d (limit %d)", len(truncBody), 20+len(suffix))
	}
	if !strings.HasSuffix(truncBody, suffix) {
		t.Errorf("truncated body should end with truncation marker: %q", truncBody)
	}
	_ = captured2
}

func TestSanitizeBody_NonJSONPassthrough(t *testing.T) {
	raw := "plain text body with no json"
	sanitized, _, _, _ := SanitizeBody(raw, DefaultRedactionOptions())
	if sanitized != raw {
		t.Errorf("non-JSON body should pass through unchanged, got %q", sanitized)
	}
}

func TestMaskPhoneAndEmail(t *testing.T) {
	if got := MaskPhone("09123456789"); strings.Contains(got, "34567") {
		t.Errorf("phone masked incorrectly: %q", got)
	}
	if got := MaskEmail("alireza@gmail.com"); strings.Contains(got, "lireza") {
		t.Errorf("email masked incorrectly: %q", got)
	}
	if got := MaskRecipient("09123456789", "sms"); strings.Contains(got, "12345") {
		t.Errorf("recipient masked incorrectly: %q", got)
	}
}

func TestContentHashAndPreview(t *testing.T) {
	h := ContentHash("hello")
	if len(h) != 64 {
		t.Errorf("hash should be 64 hex chars, got %d", len(h))
	}
	if ContentHash("hello") != ContentHash("hello") {
		t.Errorf("hash should be deterministic")
	}
	p := TruncatePreview("hello world", 5)
	if p != "hello…" {
		t.Errorf("preview truncation wrong: %q", p)
	}
}

func TestSanitizeBody_JSONValidityPreserved(t *testing.T) {
	body := `{"apiKey":"k","items":[{"code":"1"},{"name":"x"}]}`
	sanitized, _, _, _ := SanitizeBody(body, DefaultRedactionOptions())
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(sanitized), &parsed); err != nil {
		t.Fatalf("sanitized body is not valid JSON: %v (%q)", err, sanitized)
	}
}
