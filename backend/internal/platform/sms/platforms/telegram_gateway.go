package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/minisource/notifier/internal/platform/telegram"
)

// TelegramGatewayClientConfig carries the runtime settings for the Telegram
// Gateway adapter. It is injected by the caller (sms.NewClientFromConfig reads
// the process config and passes it here) so the platform layer stays decoupled
// from app config and can be pointed at a fake server in tests.
type TelegramGatewayClientConfig struct {
	BaseURL          string
	RequestTimeout   time.Duration
	ConnectTimeout   time.Duration
	MaxResponseBytes int64
	TTL              int    // code TTL seconds (30..3600); invalid values fall back to 120
	CheckPhone       string // optional E.164 phone used for live health checks
}

// TelegramGatewayClient adapts the OFFICIAL Telegram Gateway API
// (gatewayapi.telegram.org) into the existing SMS provider abstraction so OTP
// codes can be delivered via sendVerificationMessage.
//
// The token is the provider row's apiKey (stored encrypted in SecretConfig) or,
// when the row has none, the process-level TELEGRAM_GATEWAY_API_TOKEN. The OTP
// code is NEVER logged, persisted, or included in error messages.
type TelegramGatewayClient struct {
	client     *telegram.Client
	baseURL    string
	ttl        int
	checkPhone string
}

// ErrTelegramCheckNotConfigured is returned by Check when no health-check phone
// is configured — the token may be valid, but a live probe is impossible. The
// health check maps it to "degraded" (like kavenegar's ErrNoSenderLine), never a
// fabricated healthy or a hard "down".
var ErrTelegramCheckNotConfigured = errors.New("telegram gateway: TELEGRAM_GATEWAY_CHECK_PHONE is not set — cannot run a live health check")

// GetTelegramGatewayClient builds the Telegram Gateway client from an injected
// config (token = apiKey or process token fallback).
func GetTelegramGatewayClient(apiToken string, cfg TelegramGatewayClientConfig) (*TelegramGatewayClient, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = telegram.DefaultBaseURL
	}
	client, err := telegram.NewClient(telegram.Options{
		Token:            apiToken,
		BaseURL:          baseURL,
		RequestTimeout:   cfg.RequestTimeout,
		ConnectTimeout:   cfg.ConnectTimeout,
		MaxResponseBytes: cfg.MaxResponseBytes,
	})
	if err != nil {
		return nil, err
	}
	ttl := cfg.TTL
	if ttl < 30 || ttl > 3600 {
		ttl = 120
	}
	return &TelegramGatewayClient{
		client:     client,
		baseURL:    baseURL,
		ttl:        ttl,
		checkPhone: cfg.CheckPhone,
	}, nil
}

// SendMessage implements SmsClient. It sends an OTP verification message via
// the official sendVerificationMessage operation. The code is read from the
// "code" / "token" / "message" params (MiniSource owns the OTP — the adapter
// never uses provider-generated code_length). TTL may be overridden per send
// via the "ttl" param.
func (t *TelegramGatewayClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	if len(targetPhoneNumber) == 0 {
		return fmt.Errorf("no target phone number provided")
	}
	code := param["code"]
	if code == "" {
		code = param["token"]
	}
	if code == "" {
		code = param["message"]
	}
	if code == "" {
		return fmt.Errorf("telegram gateway: OTP code is required (pass code/token param)")
	}

	ttl := t.ttl
	if raw, ok := param["ttl"]; ok && raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 30 && n <= 3600 {
			ttl = n
		}
	}

	for _, phone := range targetPhoneNumber {
		req := telegram.SendVerificationMessageRequest{
			PhoneNumber: phone,
			Code:        code,
			TTL:         ttl,
		}
		if _, err := t.client.SendVerificationMessage(context.Background(), req); err != nil {
			// Mask the phone in surfaced errors (OTP spec: no full phone
			// numbers in ordinary logs/errors).
			return fmt.Errorf("telegram gateway send failed for %s: %w", maskPhone(phone), err)
		}
	}
	return nil
}

// Check implements HealthCheckable. It performs a live checkSendAbility probe
// against the configured check phone (TELEGRAM_GATEWAY_CHECK_PHONE) to verify
// token validity and reachability. Without a check phone it returns
// ErrTelegramCheckNotConfigured so the health check reports "degraded" instead
// of fabricating a healthy status or a hard failure.
func (t *TelegramGatewayClient) Check(ctx context.Context) error {
	if t.checkPhone == "" {
		return ErrTelegramCheckNotConfigured
	}
	_, err := t.client.CheckSendAbility(ctx, telegram.CheckSendAbilityRequest{PhoneNumber: t.checkPhone})
	return err
}

// DescribeRequest implements RequestDescriber. It returns the exact outbound
// request for a send with the phone masked and the OTP code redacted — the
// same markers the provider-attempt logger recognizes. The token is never part
// of the URL or body.
func (t *TelegramGatewayClient) DescribeRequest(param map[string]string, targetPhoneNumber ...string) (method, requestURL, requestBody string) {
	if len(targetPhoneNumber) == 0 {
		return "", "", ""
	}
	code := param["code"]
	if code == "" {
		code = param["token"]
	}
	if code == "" {
		code = param["message"]
	}
	if code == "" {
		return "", "", ""
	}
	ttl := t.ttl
	if raw, ok := param["ttl"]; ok && raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 30 && n <= 3600 {
			ttl = n
		}
	}
	payload := map[string]interface{}{
		"phone_number": maskPhone(targetPhoneNumber[0]),
		"code":         redactionMarker,
		"ttl":          ttl,
	}
	body, _ := json.Marshal(payload)
	return "POST", t.baseURL + "/sendVerificationMessage", string(body)
}

var _ SmsClient = (*TelegramGatewayClient)(nil)
var _ HealthCheckable = (*TelegramGatewayClient)(nil)
var _ RequestDescriber = (*TelegramGatewayClient)(nil)
