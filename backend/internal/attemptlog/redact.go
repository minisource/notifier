package attemptlog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// Redaction markers — surfaced verbatim in the UI so operators always know a
// value was intentionally removed, masked, or truncated.
const (
	MarkerRedacted   = "[REDACTED]"
	MarkerNotCaptured = "[NOT CAPTURED]"
)

// sensitiveHeaderKeys are HTTP header names whose values must never be stored.
var sensitiveHeaderKeys = map[string]bool{
	"authorization":     true,
	"proxy-authorization": true,
	"cookie":            true,
	"set-cookie":        true,
	"x-api-key":         true,
	"api-key":           true,
	"x-auth-token":      true,
	"x-access-token":    true,
	"x-csrf-token":      true,
	"x-real-ip":         true,
	"x-forwarded-for":   true,
	"x-forwarded-host":  true,
	"x-request-id":      false, // safe, keep
}

// sensitiveQueryKeys are query parameter names whose values must be redacted.
var sensitiveQueryKeys = map[string]bool{
	"api_key": true, "apikey": true, "key": true, "token": true, "access_token": true,
	"auth": true, "signature": true, "sig": true, "password": true, "secret": true,
	"client_secret": true, "refresh_token": true, "code": true, "otp": true,
}

// sensitiveBodyKeys are JSON body keys (case-insensitive substring match) that
// must be redacted when capturing request/response bodies.
//
// NOTE: "code" is deliberately NOT substring-matched here: substring matching
// would also redact benign diagnostic keys like statusCode/errorCode and destroy
// exactly the operational value this feature exists to keep. "code" (the exact
// Kavenegar OTP param key) is handled by exact matching in isSensitiveBodyKey.
var sensitiveBodyKeys = []string{
	"apikey", "api_key", "api-key", "accesskey", "access_key", "access-key",
	"secret", "password", "passwd", "token", "authorization", "bearer",
	"clientsecret", "client_secret", "client-secret", "refresh_token", "refresh-token",
	"otp", "verifycode", "verification_code", "verificationcode", "privatekey", "private_key",
}

// exactSensitiveBodyKeys are matched on the exact lowercased key only.
var exactSensitiveBodyKeys = map[string]bool{
	"code": true, // OTP/code values are sensitive in message payloads
}

// RedactionOptions controls body capture bounds.
type RedactionOptions struct {
	// MaxBodyBytes is the max captured size of a request/response body before
	// truncation. 0 = no body capture.
	MaxBodyBytes int
	// MaxPreviewChars bounds the stored message preview. 0 = no preview.
	MaxPreviewChars int
}

// DefaultRedactionOptions returns safe defaults.
func DefaultRedactionOptions() RedactionOptions {
	return RedactionOptions{MaxBodyBytes: 8192, MaxPreviewChars: 200}
}

// SanitizeHeaderMap returns a copy of headers with sensitive values replaced
// by MarkerRedacted. Safe keys like X-Request-Id are preserved.
func SanitizeHeaderMap(headers map[string]string) map[string]string {
	if headers == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		lk := strings.ToLower(k)
		if sensitiveHeaderKeys[lk] {
			out[k] = MarkerRedacted
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return map[string]string{}
	}
	return out
}

// SanitizeQueryString returns the query string with sensitive parameter values
// redacted. Empty result returns "".
func SanitizeQueryString(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return MarkerRedacted // can't parse — don't risk leaking secrets
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		for _, v := range values[k] {
			if sensitiveQueryKeys[strings.ToLower(k)] {
				parts = append(parts, url.QueryEscape(k)+"="+MarkerRedacted)
			} else {
				parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
			}
		}
	}
	return strings.Join(parts, "&")
}

// SanitizeURL redacts query strings, credentials, and key-like path segments
// (e.g. Kavenegar embeds the API key in the URL path: /v1/{APIKEY}/...).
func SanitizeURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.User != nil {
		u.User = url.User(MarkerRedacted)
	}
	u.Path = redactPathSegments(u.Path)
	u.RawQuery = SanitizeQueryString(u.RawQuery)
	return u.String()
}

// redactPathSegments redacts path segments that look like embedded credentials
// (long alphanumeric runs, e.g. API keys, signatures). Short safe segments such
// as template names ("verify", "lookup.json") are preserved.
func redactPathSegments(path string) string {
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	changed := false
	for i, seg := range parts {
		if seg == "" || seg == "." || seg == ".." {
			continue
		}
		if looksLikeCredential(seg) {
			parts[i] = MarkerRedacted
			changed = true
		}
	}
	if !changed {
		return path
	}
	return strings.Join(parts, "/")
}

// looksLikeCredential reports whether a URL path segment resembles an embedded
// secret: pure alphanumeric run of 12+ chars with mixed case or digits (API
// keys, signatures) — while allowing short words and dotted filenames through.
func looksLikeCredential(seg string) bool {
	if len(seg) < 12 {
		return false
	}
	if strings.ContainsAny(seg, ".-_") {
		return false
	}
	hasLetter, hasDigit := false, false
	for _, r := range seg {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			return false
		}
	}
	return hasLetter && hasDigit
}

// SanitizeBody redacts sensitive keys from a JSON body and truncates it.
// Returns sanitized body, truncated flag, original size, captured size.
func SanitizeBody(raw string, opts RedactionOptions) (string, bool, int, int) {
	if raw == "" {
		return "", false, 0, 0
	}
	original := len([]byte(raw))
	sanitized := redactJSONString(raw)
	captured := len([]byte(sanitized))
	truncated := false

	if opts.MaxBodyBytes > 0 && captured > opts.MaxBodyBytes {
		truncated = true
		sanitized = sanitized[:opts.MaxBodyBytes] + "\n...[TRUNCATED]"
		captured = len([]byte(sanitized))
	}
	return sanitized, truncated, original, captured
}

// redactJSONString attempts to redact a JSON body; falls back to raw text when
// the body is not JSON (still bounded later by the caller).
func redactJSONString(raw string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return raw
	}
	redacted := redactValue(data)
	b, err := json.Marshal(redacted)
	if err != nil {
		return raw
	}
	return string(b)
}

// redactValue recursively replaces sensitive keys with MarkerRedacted.
func redactValue(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(t))
		for k, val := range t {
			if isSensitiveBodyKey(k) {
				out[k] = MarkerRedacted
			} else {
				out[k] = redactValue(val)
			}
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, item := range t {
			out[i] = redactValue(item)
		}
		return out
	default:
		return v
	}
}

func isSensitiveBodyKey(key string) bool {
	lower := strings.ToLower(key)
	if exactSensitiveBodyKeys[lower] {
		return true
	}
	for _, s := range sensitiveBodyKeys {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

// MaskPhone masks a phone number (keeps first 2 and last 2 chars).
func MaskPhone(phone string) string {
	if phone == "" || len(phone) < 6 {
		return phone
	}
	return phone[:2] + strings.Repeat("*", len(phone)-4) + phone[len(phone)-2:]
}

// MaskEmail masks an email address (keeps first char of local part).
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}
	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return "***@***"
	}
	local := email[:at]
	domain := email[at+1:]
	if len(local) <= 1 {
		return local + "***@" + domain
	}
	return local[:1] + strings.Repeat("*", len(local)-1) + "@" + domain
}

// MaskRecipient masks a recipient value heuristically (phone vs email).
func MaskRecipient(value, channel string) string {
	if value == "" {
		return ""
	}
	if strings.Contains(value, "@") {
		return MaskEmail(value)
	}
	return MaskPhone(value)
}

// ContentHash returns a SHA-256 hex digest of the message content. Used to
// correlate messages without persisting the content itself.
func ContentHash(content string) string {
	if content == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// TruncatePreview bounds the stored message preview.
func TruncatePreview(content string, max int) string {
	if content == "" {
		return ""
	}
	if max <= 0 {
		return ""
	}
	r := []rune(content)
	if len(r) <= max {
		return content
	}
	return string(r[:max]) + "…"
}

// FormatBytes returns a human-readable byte size.
func FormatBytes(n int) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}
