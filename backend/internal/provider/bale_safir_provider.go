package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BaleSafirConfig holds the configuration for the Bale Safir SMS provider
type BaleSafirConfig struct {
	// V3 API (messaging)
	AccessKey string `json:"accessKey"`
	BotID     int64  `json:"botId"`

	// V2 API (OTP) — client_credentials OAuth2
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`

	// Default mode: "otp" or "message"
	Mode string `json:"mode"`
}

// BaleSafirProvider sends SMS via Bale Safir API (سرویس سفیر بله)
type BaleSafirProvider struct {
	name   string
	config *BaleSafirConfig
	client *http.Client

	// OAuth token cache for V2 API
	otpToken     string
	otpTokenExp  time.Time
}

// NewBaleSafirProvider creates a new Bale Safir provider from config JSON
func NewBaleSafirProvider(name, configJSON string) (*BaleSafirProvider, error) {
	var cfg BaleSafirConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid Bale Safir config: %w", err)
	}
	return NewBaleSafirProviderFromConfig(name, &cfg), nil
}

// NewBaleSafirProviderFromConfig creates a Bale Safir provider from parsed config
func NewBaleSafirProviderFromConfig(name string, cfg *BaleSafirConfig) *BaleSafirProvider {
	if cfg.Mode == "" {
		cfg.Mode = "message"
	}
	return &BaleSafirProvider{
		name:   name,
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *BaleSafirProvider) Channel() Channel {
	return ChannelSMS
}

func (p *BaleSafirProvider) Name() string {
	return p.name
}

// Send delivers an SMS via Bale Safir
func (p *BaleSafirProvider) Send(ctx context.Context, msg *Message) (*SendResult, error) {
	if msg.To == "" {
		return NewFailureResult(p.Name(), p.Channel(), ErrorInvalidRecipient, "phone number is required"), nil
	}

	// Determine mode: metadata overrides config default
	mode := p.config.Mode
	if msg.Metadata != nil {
		if m, ok := msg.Metadata["mode"]; ok && m != "" {
			mode = m
		}
	}

	var err error
	switch mode {
	case "otp":
		err = p.sendOTP(ctx, msg.To, msg.Body)
	default:
		err = p.sendMessage(ctx, msg.To, msg.Body)
	}

	if err != nil {
		code := classifyBaleError(err)
		return NewFailureResult(p.Name(), p.Channel(), code, err.Error()), nil
	}

	providerID := fmt.Sprintf("bale_safir_%s_%s", p.name, msg.NotificationID)
	return NewSuccessResult(p.Name(), p.Channel(), providerID), nil
}

// sendMessage sends a text message via Safir V3 API
func (p *BaleSafirProvider) sendMessage(ctx context.Context, to, body string) error {
	if p.config.AccessKey == "" {
		return fmt.Errorf("Safir accessKey is not configured for V3 messaging")
	}
	if p.config.BotID == 0 {
		return fmt.Errorf("Safir botId is not configured for V3 messaging")
	}

	payload := map[string]interface{}{
		"bot_id":       p.config.BotID,
		"phone_number": to,
		"message_data": map[string]interface{}{
			"message": map[string]string{
				"text": body,
			},
		},
	}

	return p.safirV3Request(ctx, "send_message", payload)
}

// sendOTP sends an OTP code via Safir V2 API
func (p *BaleSafirProvider) sendOTP(ctx context.Context, to, code string) error {
	if p.config.ClientID == "" || p.config.ClientSecret == "" {
		return fmt.Errorf("Safir clientId/clientSecret are not configured for OTP")
	}

	if err := p.ensureOTPToken(ctx); err != nil {
		return fmt.Errorf("failed to get Safir OTP token: %w", err)
	}

	payload := map[string]string{
		"phone": to,
		"otp":   code,
	}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://safir.bale.ai/api/v2/send_otp", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create OTP request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.otpToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("Safir OTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Safir OTP error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// safirV3Request makes a request to the Safir V3 API
func (p *BaleSafirProvider) safirV3Request(ctx context.Context, endpoint string, payload interface{}) error {
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://safir.bale.ai/api/v3/"+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Safir V3 request: %w", err)
	}
	req.Header.Set("api-access-key", p.config.AccessKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("Safir V3 request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		var safirErr struct {
			Type    int    `json:"type"`
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &safirErr) == nil && safirErr.Message != "" {
			return fmt.Errorf("Safir V3 error [%d]: %s", safirErr.Code, safirErr.Message)
		}
		return fmt.Errorf("Safir V3 error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// ensureOTPToken gets or refreshes the OAuth2 token for V2 API
func (p *BaleSafirProvider) ensureOTPToken(ctx context.Context) error {
	if p.otpToken != "" && time.Now().Before(p.otpTokenExp) {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("scope", "read")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://safir.bale.ai/api/v2/auth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse OTP token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return fmt.Errorf("Safir OTP auth returned empty token: %s", string(body))
	}

	p.otpToken = tokenResp.AccessToken
	if tokenResp.ExpiresIn > 0 {
		p.otpTokenExp = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
	} else {
		p.otpTokenExp = time.Now().Add(55 * time.Minute)
	}
	return nil
}

func classifyBaleError(err error) ProviderErrorCode {
	errStr := err.Error()
	switch {
	case containsAny(errStr, "timed out", "timeout", "context deadline"):
		return ErrorTimeout
	case containsAny(errStr, "rate limit", "too many requests"):
		return ErrorRateLimited
	case containsAny(errStr, "connection refused", "no such host", "network"):
		return ErrorNetworkError
	case containsAny(errStr, "not configured", "invalid config", "accessKey", "access_key", "clientId", "client_id", "clientSecret", "client_secret", "botId", "bot_id"):
		return ErrorInvalidConfig
	case containsAny(errStr, "phone", "recipient", "invalid number", "phone_number"):
		return ErrorInvalidRecipient
	case containsAny(errStr, "403", "401", "unauthorized", "forbidden", "permission"):
		return ErrorInvalidConfig
	case containsAny(errStr, "503", "502", "service unavailable"):
		return ErrorServiceUnavailable
	case containsAny(errStr, "400", "bad request"):
		return ErrorInvalidMessage
	default:
		return ErrorProviderError
	}
}
