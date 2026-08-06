// Package telegram implements the OFFICIAL Telegram Gateway API
// (https://core.telegram.org/gateway/api). Telegram Gateway is NOT the Bot API:
// it sends real verification messages via sendVerificationMessage and related
// operations. This client only uses the official contract — no unofficial
// endpoints, no bot tokens, no guessed fields.
//
// Security rules enforced here:
//   - The API token is sent as `Authorization: Bearer <token>` and is NEVER
//     included in any error string, log line, or returned value.
//   - All timeouts are bounded (connect + overall request + response cap).
//   - Response bodies are size-capped.
//   - OTP/code values are never returned in error messages.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// DefaultBaseURL is the official Telegram Gateway API root.
const DefaultBaseURL = "https://gatewayapi.telegram.org"

// Official operation names (also used as the endpoint path suffix).
const (
	OpCheckSendAbility        = "checkSendAbility"
	OpSendVerificationMessage = "sendVerificationMessage"
	OpCheckVerificationStatus = "checkVerificationStatus"
	OpRevokeVerificationMessage = "revokeVerificationMessage"
)

// Client is a minimal, secure Telegram Gateway API client.
type Client struct {
	token            string
	baseURL          string
	httpClient       *http.Client
	maxResponseBytes int64
}

// Options configures the client.
type Options struct {
	Token            string
	BaseURL          string        // default DefaultBaseURL
	RequestTimeout   time.Duration // overall request deadline
	ConnectTimeout   time.Duration // TCP connect + TLS handshake
	MaxResponseBytes int64
}

// NewClient creates the client. BaseURL is validated (https required unless a
// test override is explicitly supplied by the caller — production config is
// validated at startup separately).
func NewClient(opts Options) (*Client, error) {
	if opts.Token == "" {
		return nil, errors.New("telegram gateway: api token is required")
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(baseURL, "https://") && !strings.HasPrefix(baseURL, "http://") {
		return nil, fmt.Errorf("telegram gateway: invalid base URL %q (must be http(s)://)", baseURL)
	}
	connectTimeout := opts.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 5 * time.Second
	}
	requestTimeout := opts.RequestTimeout
	if requestTimeout <= 0 {
		requestTimeout = 15 * time.Second
	}
	maxBody := opts.MaxResponseBytes
	if maxBody <= 0 {
		maxBody = 1 << 20
	}
	transport := &http.Transport{
		DialContext: (&net.Dialer{Timeout: connectTimeout}).DialContext,
		// TLS handshake is bounded by connectTimeout as well.
		TLSHandshakeTimeout: connectTimeout,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
	}
	return &Client{
		token:            opts.Token,
		baseURL:          baseURL,
		httpClient:       &http.Client{Transport: transport, Timeout: requestTimeout},
		maxResponseBytes: maxBody,
	}, nil
}

// ---- Official request/response types (field names per core.telegram.org/gateway/api) ----

// CheckSendAbilityRequest is the official request for checkSendAbility.
type CheckSendAbilityRequest struct {
	PhoneNumber string `json:"phone_number"`
}

// SendVerificationMessageRequest is the official request for sendVerificationMessage.
// Either Code (application-provided, 4-8 digits) or CodeLength (provider-generated)
// is used; TTL must be within 30..3600 seconds; RequestID is the (optional)
// identifier returned by a previous checkSendAbility (making the send free).
type SendVerificationMessageRequest struct {
	PhoneNumber   string `json:"phone_number"`
	RequestID     string `json:"request_id,omitempty"`
	SenderUsername string `json:"sender_username,omitempty"`
	Code          string `json:"code,omitempty"`
	CodeLength    int    `json:"code_length,omitempty"`
	CallbackURL   string `json:"callback_url,omitempty"`
	Payload       string `json:"payload,omitempty"`
	TTL           int    `json:"ttl,omitempty"`
}

// CheckVerificationStatusRequest is the official request for checkVerificationStatus.
type CheckVerificationStatusRequest struct {
	RequestID string `json:"request_id"`
	Code      string `json:"code,omitempty"` // user-entered code (optional)
}

// RevokeVerificationMessageRequest is the official request for revokeVerificationMessage.
type RevokeVerificationMessageRequest struct {
	RequestID string `json:"request_id"`
}

// DeliveryStatus is the official delivery status object.
type DeliveryStatus struct {
	Status    string `json:"status"` // sent | delivered | read | expired | revoked
	UpdatedAt int64  `json:"updated_at"`
}

// VerificationStatus is the official verification status object.
type VerificationStatus struct {
	Status      string `json:"status"` // code_valid | code_invalid | code_max_attempts_exceeded | expired
	UpdatedAt   int64  `json:"updated_at"`
	CodeEntered string `json:"code_entered,omitempty"` // sensitive — never persisted/logged
}

// RequestStatus is the official result object returned by the API.
type RequestStatus struct {
	RequestID         string             `json:"request_id"`
	PhoneNumber       string             `json:"phone_number"`
	RequestCost       float64            `json:"request_cost"`
	IsRefunded        bool               `json:"is_refunded"`
	RemainingBalance  float64            `json:"remaining_balance"`
	DeliveryStatus    *DeliveryStatus    `json:"delivery_status"`
	VerificationStatus *VerificationStatus `json:"verification_status"`
	Payload           string             `json:"payload"`
}

// envelope is the official response wrapper: ok + result | error.
type envelope struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

// ---- Operations ----

// CheckSendAbility checks whether a phone number can receive a verification
// message right now. Returns the official result (request_id + cost).
func (c *Client) CheckSendAbility(ctx context.Context, req CheckSendAbilityRequest) (*RequestStatus, error) {
	return c.do(ctx, OpCheckSendAbility, req)
}

// SendVerificationMessage sends a verification message. Returns the official
// result (request_id, delivery/verification status, cost).
func (c *Client) SendVerificationMessage(ctx context.Context, req SendVerificationMessageRequest) (*RequestStatus, error) {
	return c.do(ctx, OpSendVerificationMessage, req)
}

// CheckVerificationStatus checks the status of a verification request,
// optionally validating a user-entered code. Returns the official result.
func (c *Client) CheckVerificationStatus(ctx context.Context, req CheckVerificationStatusRequest) (*RequestStatus, error) {
	return c.do(ctx, OpCheckVerificationStatus, req)
}

// RevokeVerificationMessage revokes a previously sent verification message.
func (c *Client) RevokeVerificationMessage(ctx context.Context, req RevokeVerificationMessageRequest) (*RequestStatus, error) {
	return c.do(ctx, OpRevokeVerificationMessage, req)
}

// do executes one official operation. The request body may contain the OTP —
// it is sent to Telegram but is never logged and never returned in errors.
// The Authorization header is never logged.
func (c *Client) do(ctx context.Context, operation string, payload interface{}) (*RequestStatus, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("telegram gateway: marshal %s request: %w", operation, err)
	}
	endpoint := c.baseURL + "/" + operation
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("telegram gateway: build %s request: %w", operation, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, classifyTransportError(operation, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, c.maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("telegram gateway: read %s response: %w", operation, err)
	}

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("telegram gateway: malformed %s response (HTTP %d)", operation, resp.StatusCode)
	}
	// Non-2xx is never success, even if the body claims ok:true (proxy/HTTP
	// level failures must not masquerade as provider success).
	if resp.StatusCode < http.StatusOK || resp.StatusCode > 299 {
		return nil, newAPIError(operation, env.Error, resp.StatusCode)
	}
	if !env.OK {
		// Official error model: {ok:false, error:"ERROR_CODE"}.
		return nil, newAPIError(operation, env.Error, resp.StatusCode)
	}

	var result RequestStatus
	if len(env.Result) > 0 {
		if err := json.Unmarshal(env.Result, &result); err != nil {
			return nil, fmt.Errorf("telegram gateway: malformed %s result", operation)
		}
	}
	return &result, nil
}

// ---- Error model ----

// apiError carries the official Telegram error code plus the HTTP status.
// The code string (e.g. ACCESS_TOKEN_INVALID) is safe to surface; the token
// itself is never part of it.
type apiError struct {
	Operation string
	Code      string
	HTTP      int
}

func (e *apiError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("telegram gateway %s: %s (HTTP %d)", e.Operation, e.Code, e.HTTP)
	}
	return fmt.Sprintf("telegram gateway %s: HTTP %d", e.Operation, e.HTTP)
}

func newAPIError(operation, code string, httpStatus int) error {
	return &apiError{Operation: operation, Code: code, HTTP: httpStatus}
}

// ErrorCode returns the official Telegram error code (empty when the error is
// not a Telegram API error).
func ErrorCode(err error) string {
	var ae *apiError
	if errors.As(err, &ae) {
		return ae.Code
	}
	return ""
}

// classifyTransportError maps connect/timeout/network failures to a safe error.
// Timeout errors wrap context.DeadlineExceeded so IsTimeout can rely on
// errors.Is instead of fragile string matching.
func classifyTransportError(operation string, err error) error {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("telegram gateway %s: timeout (%w)", operation, context.DeadlineExceeded)
	}
	return fmt.Errorf("telegram gateway %s: network error (%w)", operation, err)
}

// IsTimeout reports whether the error is a transport timeout. nil is never a
// timeout.
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, context.DeadlineExceeded)
}

// NormalizedErrorKind classifies a Telegram Gateway error into the domain's
// normalized categories (authentication, rate_limited, insufficient_balance,
// invalid_recipient, invalid_request, provider, timeout, network, unknown).
func NormalizedErrorKind(err error) string {
	if IsTimeout(err) {
		return "timeout"
	}
	code := ErrorCode(err)
	switch code {
	case "ACCESS_TOKEN_INVALID", "ACCESS_TOKEN_EXPIRED", "UNAUTHORIZED":
		return "authentication"
	case "RATE_LIMITED", "TOO_MANY_REQUESTS":
		return "rate_limited"
	case "INSUFFICIENT_BALANCE":
		return "insufficient_balance"
	case "PHONE_NUMBER_INVALID", "INVALID_PHONE_NUMBER":
		return "invalid_recipient"
	case "REQUEST_NOT_FOUND", "INVALID_REQUEST_ID":
		return "invalid_request"
	case "INTERNAL_ERROR", "SERVICE_UNAVAILABLE":
		return "provider"
	default:
		if err == nil {
			return "success"
		}
		return "provider"
	}
}
