package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kavenegar/kavenegar-go"
)

const defaultKavenegarBaseURL = "https://api.kavenegar.com"

type KavenegarClient struct {
	apiKey   string
	template string
	sender   string
	core     *kavenegar.Kavenegar
	// baseURL overrides the API root (used by tests; defaults to the real API).
	baseURL string
}

func GetKavenegarClient(accessKey string, template string, sender string) (*KavenegarClient, error) {
	client := kavenegar.New(accessKey)

	kavenegarClient := &KavenegarClient{
		core:     client,
		apiKey:   accessKey,
		template: template,
		sender:   sender,
	}

	return kavenegarClient, nil
}

// KavenegarAccountInfo is the sanitized account information returned by the
// account/info.json endpoint. Only safe operational fields are exposed — the
// API key is never included, and the raw URL is never persisted.
type KavenegarAccountInfo struct {
	RemainCredit    *float64  // remaining credit balance in Rial (monetary amount, not a message count)
	ExpireDate      *time.Time
	AccountType     string
	SignatureText   string
	IsVerifySend    bool
	ProviderStatus  int
	ProviderMessage string
}

// AccountInfo calls the Kavenegar account/info.json endpoint and returns a
// sanitized snapshot of the account. The API key is embedded in the request
// URL path but is never returned, persisted, or logged. A refresh failure is
// normalized into a typed error kind (authentication/rate_limited/timeout/
// network/malformed) via NormalizeKavenegarBalanceError so the balance service
// can classify it without leaking secrets.
func (k *KavenegarClient) AccountInfo(ctx context.Context) (*KavenegarAccountInfo, error) {
	if k.apiKey == "" {
		return nil, fmt.Errorf("kavenegar api key is not configured")
	}

	endpoint := fmt.Sprintf("%s/v1/%s/account/info.json", k.base(), k.apiKey)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("kavenegar account info request failed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kavenegar account info request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("kavenegar account info read failed: %w", err)
	}

	var payload struct {
		Return struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		} `json:"return"`
		// Entries is json.RawMessage because Kavenegar returns an ARRAY on
		// success but an empty OBJECT ({}) in some account states — unmarshal
		// must tolerate both instead of failing the whole refresh.
		Entries json.RawMessage `json:"entries"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("kavenegar account info: malformed response: %w", err)
	}

	// Kavenegar reports API errors with HTTP 200 and the error inside the body
	// (return.status/return.message). Both transport status and body status must
	// be checked.
	if resp.StatusCode != http.StatusOK || (payload.Return.Status != 0 && payload.Return.Status != 200) {
		return nil, newKavenegarBalanceError(payload.Return.Status, payload.Return.Message, resp.StatusCode)
	}

	info := &KavenegarAccountInfo{
		AccountType:     "",
		ProviderStatus:  payload.Return.Status,
		ProviderMessage: payload.Return.Message,
	}

	// Kavenegar returns `entries` as an ARRAY for some account types and as an
	// OBJECT ({"remaincredit":..., "expiredate":"...", "type":...}) for
	// others. Both must be parsed. Field types also vary: expiredate is a Unix
	// number in some responses and a numeric STRING in others. The generic map
	// approach tolerates every variant without failing the refresh.
	if len(payload.Entries) > 0 && payload.Entries[0] != 'n' {
		var objects []map[string]interface{}
		if payload.Entries[0] == '[' {
			var arr []map[string]interface{}
			if err := json.Unmarshal(payload.Entries, &arr); err != nil {
				return nil, fmt.Errorf("kavenegar account info: malformed entries: %w", err)
			}
			objects = arr
		} else {
			var obj map[string]interface{}
			if err := json.Unmarshal(payload.Entries, &obj); err != nil {
				return nil, fmt.Errorf("kavenegar account info: malformed entries: %w", err)
			}
			objects = []map[string]interface{}{obj}
		}
		if len(objects) > 0 {
			e := objects[0]
			if rc, ok := toFloat64(e["remaincredit"]); ok {
				info.RemainCredit = &rc
			}
			if ts, ok := toInt64(e["expiredate"]); ok && ts > 0 {
				t := time.Unix(ts, 0)
				info.ExpireDate = &t
			}
			if s, ok := e["type"].(string); ok {
				info.AccountType = s
			}
			if s, ok := e["signtext"].(string); ok {
				info.SignatureText = s
			}
			if v, ok := toInt64(e["isverifysend"]); ok {
				info.IsVerifySend = v == 1
			}
		}
	}
	return info, nil
}

// toFloat64 extracts a float64 from a JSON value that may arrive as float64,
// json.Number, or a numeric string.
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		var f float64
		if _, err := fmt.Sscanf(n, "%f", &f); err == nil {
			return f, true
		}
	}
	return 0, false
}

// toInt64 extracts an int64 from a JSON value that may arrive as float64,
// json.Number, or a numeric string (Kavenegar's expiredate is sometimes a
// string).
func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	case string:
		var i int64
		if _, err := fmt.Sscanf(n, "%d", &i); err == nil {
			return i, true
		}
	}
	return 0, false
}

// kavenegarBalanceErr normalizes a Kavenegar API error into a typed error
// whose kind can be classified by NormalizeKavenegarBalanceError. The raw
// message is kept (it is provider-localized, not a secret) and the status code
// is preserved for mapping.
type kavenegarBalanceErr struct {
	Status  int
	Message string
	HTTP    int
}

func (e *kavenegarBalanceErr) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("kavenegar rejected request: %s (status %d)", e.Message, e.Status)
	}
	return fmt.Sprintf("kavenegar account info returned HTTP %d", e.HTTP)
}

func newKavenegarBalanceError(status int, message string, httpStatus int) error {
	return &kavenegarBalanceErr{Status: status, Message: message, HTTP: httpStatus}
}

// NormalizeKavenegarBalanceError classifies a Kavenegar account-info error into
// a safe error kind + code. Error messages are sanitized (never contain the API
// key). Returns kind, code, sanitizedMessage.
func NormalizeKavenegarBalanceError(err error) (kind, code, sanitizedMessage string) {
	var be *kavenegarBalanceErr
	if errors.As(err, &be) {
		switch be.Status {
		case 1:
			return "malformed", "kavenegar_1", "Kavenegar returned an invalid API response"
		case 2, 3, 4, 5, 6:
			return "authentication", "kavenegar_" + fmt.Sprint(be.Status), "Kavenegar rejected the API key or the account is restricted"
		case 332:
			return "authentication", "kavenegar_332", "Kavenegar API key is invalid or expired"
		case 333:
			return "network", "kavenegar_333", "Kavenegar connection failed"
		case 34:
			return "rate_limited", "kavenegar_34", "Kavenegar rate limit exceeded"
		default:
			return "unknown", "kavenegar_" + fmt.Sprint(be.Status), be.Message
		}
	}
	return "network", "kavenegar_network", err.Error()
}

// ErrNoSenderLine is returned by Check when the Kavenegar account has no
// default sender line AND the provider config sets neither senderId nor a
// template. In that state every plain (sms/send.json) send is rejected with
// APIError[412] "ارسال کننده نامعتبر است", so the provider is degraded: the
// API key is valid but sends cannot succeed. The health check maps this to a
// "degraded" status with an actionable message.
var ErrNoSenderLine = fmt.Errorf("kavenegar: no sender line available for plain sends")

// Check verifies the API key and connectivity by calling the real Kavenegar
// account/info endpoint. It uses a bounded HTTP client so the caller's
// context timeout is actually honored (unlike k.core.Account.Info(), which
// does not accept a context). It fails when the API key is invalid or the
// service is unreachable.
//
// When the key is valid it additionally inspects account/config to detect the
// "key works but every send would be rejected with 412" state (no default
// sender line + no senderId/template configured) and returns ErrNoSenderLine.
func (k *KavenegarClient) Check(ctx context.Context) error {
	if k.apiKey == "" {
		return fmt.Errorf("kavenegar api key is not configured")
	}

	endpoint := fmt.Sprintf("%s/v1/%s/account/info.json", k.base(), k.apiKey)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("kavenegar account info request failed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("kavenegar account info request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("kavenegar account info read failed: %w", err)
	}

	// Kavenegar reports API errors with HTTP 200 and the error inside the body
	// (return.status/return.message), so both transport status and body status
	// must be checked.
	var apiErr struct {
		Return struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		} `json:"return"`
	}
	_ = json.Unmarshal(body, &apiErr)
	if resp.StatusCode != http.StatusOK || (apiErr.Return.Status != 0 && apiErr.Return.Status != 200) {
		if apiErr.Return.Message != "" {
			return fmt.Errorf("kavenegar rejected API key: %s (status %d)", apiErr.Return.Message, apiErr.Return.Status)
		}
		return fmt.Errorf("kavenegar account info returned HTTP %d", resp.StatusCode)
	}

	// The key is valid. Check whether plain sends are actually possible: if the
	// account has no default sender line and this provider sets neither a
	// senderId nor a template, every sms/send.json call is rejected with 412.
	// Template (lookup) sends would still work, so report this as degraded.
	defaultSender, cfgErr := k.defaultSender(ctx, client)
	if cfgErr != nil {
		// Config endpoint unavailable — the key itself is fine, do not fail the
		// whole check over a secondary probe.
		return nil
	}
	if defaultSender == "" && k.sender == "" && k.template == "" {
		return ErrNoSenderLine
	}

	return nil
}

// base returns the API root to use, allowing tests to point at a local server.
func (k *KavenegarClient) base() string {
	if k.baseURL != "" {
		return k.baseURL
	}
	return defaultKavenegarBaseURL
}

// defaultSender fetches account/config.json and returns the account's
// configured default sender line (may be empty when none is set).
func (k *KavenegarClient) defaultSender(ctx context.Context, client *http.Client) (string, error) {
	endpoint := fmt.Sprintf("%s/v1/%s/account/config.json", k.base(), k.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	var payload struct {
		Return struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		} `json:"return"`
		Entries []struct {
			DefaultSender string `json:"defaultsender"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if payload.Return.Status != 200 || len(payload.Entries) == 0 {
		return "", fmt.Errorf("unexpected account config response")
	}
	return payload.Entries[0].DefaultSender, nil
}

// sendPlainMessage sends one free-text SMS via the Kavenegar simple message
// API (sms/send.json).
//
// The kavenegar-go SDK's Message.Send always sets the "sender" form value,
// even when the sender line is empty — Kavenegar then rejects the request with
// APIError[412] ("invalid sender"). To avoid that, when no sender line is
// configured the "sender" parameter is omitted entirely so Kavenegar falls
// back to the account's default sender line.
func (k *KavenegarClient) sendPlainMessage(sender, phoneNumber, message string) error {
	v := url.Values{}
	senderSent := sender != ""
	if senderSent {
		v.Set("sender", sender)
	}
	v.Set("receptor", phoneNumber)
	v.Set("message", message)
	_, err := k.core.Message.CreateSend(v)
	if err != nil {
		return describeKavenegarError(err, senderSent, sender)
	}
	return nil
}

// describeKavenegarError inspects the error returned by the kavenegar-go SDK
// and appends an actionable explanation for well-known API error codes. The
// SDK wraps provider rejections in *kavenegar.APIError{Status, Message} whose
// Message is already localized by Kavenegar (e.g. "ارسال کننده نامعتبر است").
// senderSent reports whether a sender line was actually included in the
// request, so the 412 hint can distinguish "no senderId configured" from
// "configured sender line is not approved".
func describeKavenegarError(err error, senderSent bool, sender string) error {
	var apiErr *kavenegar.APIError
	if !errors.As(err, &apiErr) {
		return err
	}
	switch apiErr.Status {
	case 412:
		if !senderSent {
			return fmt.Errorf("%w (no senderId configured: Kavenegar has no approved default sender line for this account — register a sender line in the Kavenegar panel and set its number in the provider's senderId field)", err)
		}
		return fmt.Errorf("%w (sender line %q rejected: the senderId is not registered or approved in the Kavenegar panel, or is not allowed for this message type — check the line number and its activation status)", err, sender)
	case 431:
		return fmt.Errorf("%w (code/token structure rejected: the value sent as the template token does not match the template's expected format — for OTP templates the code must be the exact number of digits the template requires; check the template pattern in the Kavenegar panel and the variables you pass)", err)
	case 406, 413:
		return fmt.Errorf("%w (recipient number rejected: check the phone number format, e.g. 09xxxxxxxxx or +98xxxxxxxxx)", err)
	case 401, 403:
		return fmt.Errorf("%w (API key rejected: verify the apiKey saved in the provider config)", err)
	default:
		return err
	}
}

// DescribeRequest implements providers.RequestDescriber. It returns the exact
// outbound HTTP request the adapter would send for a send: HTTP method, URL,
// and form-encoded body. The API key is redacted from the URL path, the
// recipient phone is masked, and OTP/token values are replaced with
// [REDACTED]. The provider-attempt logger uses this so operators see the REAL
// request Kavenegar receives instead of the internal params map.
// redactionMarker and maskPhone are local helpers for DescribeRequest. They
// intentionally do NOT depend on the attemptlog package: attemptlog imports
// internal/provider, and the provider package imports this platform package,
// so importing attemptlog here would create an import cycle. The markers must
// stay textually identical to attemptlog's so the UI treats them the same.
const redactionMarker = "[REDACTED]"

// maskPhone masks a phone number keeping the first two and last two chars,
// mirroring attemptlog.MaskPhone exactly so the two masks agree.
func maskPhone(phone string) string {
	if len(phone) < 6 {
		return phone
	}
	r := []rune(phone)
	return string(r[:2]) + strings.Repeat("*", len(r)-4) + string(r[len(r)-2:])
}

// looksLikeOTP reports whether a message value is OTP-shaped (short all-digit
// string, e.g. a verification code). Such values must always be redacted from
// persisted request descriptions even when no template/lookup path is used.
func looksLikeOTP(s string) bool {
	if len(s) < 4 || len(s) > 10 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (k *KavenegarClient) DescribeRequest(param map[string]string, targetPhoneNumber ...string) (method, requestURL, requestBody string) {
	if len(targetPhoneNumber) == 0 {
		return "", "", ""
	}
	method = "POST"
	phone := maskPhone(targetPhoneNumber[0])
	redactedKey := redactionMarker

	// Determine template to use (mirrors SendMessage logic)
	template := k.template
	if paramTemplate, ok := param["template"]; ok && paramTemplate != "" {
		template = paramTemplate
	}

	if template == "" {
		// Plain free-text send via sms/send.json
		message := param["message"]
		if message == "" {
			message = param["body"]
		}
		if message == "" {
			return "", "", ""
		}
		sender := k.sender
		if s, ok := param["sender"]; ok && s != "" {
			sender = s
		}
		v := url.Values{}
		if sender != "" {
			v.Set("sender", sender)
		}
		v.Set("receptor", phone)
		// OTP-shaped messages (auth codes sent as free text) are always redacted.
		if looksLikeOTP(message) {
			message = redactionMarker
		}
		v.Set("message", message)
		return method, fmt.Sprintf("https://api.kavenegar.com/v1/%s/sms/send.json", redactedKey), v.Encode()
	}

	// Lookup (verification/OTP) send via verify/lookup.json — tokens are OTP
	// secrets, so only their presence is reported, never their values.
	v := url.Values{}
	v.Set("receptor", phone)
	v.Set("template", template)
	if _, ok := param["token"]; ok {
		v.Set("token", redactedKey)
	} else if _, ok := param["code"]; ok {
		v.Set("token", redactedKey)
	} else if _, ok := param["message"]; ok {
		v.Set("token", redactedKey)
	}
	for _, t := range []string{"token2", "token3", "token10", "token20"} {
		if param[t] != "" {
			v.Set(t, redactedKey)
		}
	}
	return method, fmt.Sprintf("https://api.kavenegar.com/v1/%s/verify/lookup.json", redactedKey), v.Encode()
}

//
// When a template is available (from the provider config or the "template"
// param) it uses the Lookup API (verification/OTP messages) — Kavenegar only
// sends template name + token values, not free text.
//
// When NO template is configured, it falls back to the plain message API
// (Message.Send) so a provider configured with just an API key can still send
// a free-text SMS. The sender line comes from config SenderID or the "sender"
// param.
//
// Supported lookup params:
//   - template: Template name (optional, overrides default template) e.g., "verify", "orderPlaced"
//   - token/code: Primary token (required for lookup) - the main value like OTP code
//   - token2: Secondary token (optional) - e.g., amount, status
//   - token3: Third token (optional)
//   - token10: Token with 5 spaces allowed (optional)
//   - token20: Token with 8 spaces allowed (optional)
//
// Example templates from Kavenegar panel:
//   - "verify": OTP verification, uses %token for code
//   - "orderPlaced": Order placed, uses %token for order number, %token2 for amount
//   - "paymentSuccess": Payment success, uses %token for order, %token2 for amount
func (k *KavenegarClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	if len(targetPhoneNumber) == 0 {
		return fmt.Errorf("no target phone number provided")
	}

	// Determine template to use (param template overrides default)
	template := k.template
	if paramTemplate, ok := param["template"]; ok && paramTemplate != "" {
		template = paramTemplate
	}

	// No template configured → plain free-text send via the simple message API.
	if template == "" {
		message := param["message"]
		if message == "" {
			message = param["body"]
		}
		if message == "" {
			return fmt.Errorf("no message text provided (set a template or pass message/body)")
		}
		sender := k.sender
		if s, ok := param["sender"]; ok && s != "" {
			sender = s
		}
		for _, phoneNumber := range targetPhoneNumber {
			if err := k.sendPlainMessage(sender, phoneNumber, message); err != nil {
				return fmt.Errorf("kavenegar send failed for %s: %w", phoneNumber, err)
			}
		}
		return nil
	}

	// Get the primary token (code/token)
	token, ok := param["token"]
	if !ok {
		token, ok = param["code"]
		if !ok {
			// Fall back to "message" if neither provided
			token, ok = param["message"]
			if !ok {
				return fmt.Errorf("token, code, or message parameter is required")
			}
		}
	}

	// Guard against sending unresolved template placeholders to Kavenegar. If
	// the token still contains {{...}} markers the template variables were
	// never mapped to provider tokens (e.g. the notification was created with
	// a templateId but the variables never reached the send pipeline).
	// Kavenegar rejects such a token with APIError[431] "code structure is
	// not correct", which is misleading — fail fast with a clear message.
	// Only a truncated excerpt of the token is reported (the unrendered body
	// can contain personal data); the full value is never persisted.
	if strings.Contains(token, "{{") || strings.Contains(token, "}}") {
		excerpt := token
		if len(excerpt) > 60 {
			excerpt = excerpt[:60] + "..."
		}
		return fmt.Errorf("template variables not resolved: the token still contains placeholders (%q) — pass the template variables (e.g. code/token) when creating a template-based notification", excerpt)
	}

	// Build lookup params with optional tokens
	params := &kavenegar.VerifyLookupParam{
		Tokens: make(map[string]string),
	}
	if token2, ok := param["token2"]; ok && token2 != "" {
		params.Token2 = token2
	}
	if token3, ok := param["token3"]; ok && token3 != "" {
		params.Token3 = token3
	}
	// Token10 and Token20 are sent via the Tokens map
	if token10, ok := param["token10"]; ok && token10 != "" {
		params.Tokens["token10"] = token10
	}
	if token20, ok := param["token20"]; ok && token20 != "" {
		params.Tokens["token20"] = token20
	}

	// Send to all target phone numbers using the lookup API
	// Note: We only send template name + tokens, not full message text
	for _, phoneNumber := range targetPhoneNumber {
		if _, err := k.core.Verify.Lookup(phoneNumber, template, token, params); err != nil {
			return fmt.Errorf("kavenegar lookup failed for %s with template '%s': %w", phoneNumber, template, describeKavenegarError(err, false, ""))
		}
	}

	return nil
}
