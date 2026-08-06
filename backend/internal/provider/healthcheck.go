package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/platform/email"
	"github.com/minisource/notifier/internal/platform/push"
	"github.com/minisource/notifier/internal/platform/sms"
	smsproviders "github.com/minisource/notifier/internal/platform/sms/platforms"
)

// CheckResult is the outcome of a real health check performed against a
// provider's own API (not a simulation).
type CheckResult struct {
	ProviderID string    `json:"providerId"`
	Name       string    `json:"name"`
	Channel    string    `json:"channel"`
	Type       string    `json:"type,omitempty"`
	Status     string    `json:"status"` // healthy | degraded | down | disabled | unsupported
	Message    string    `json:"message,omitempty"`
	Error      string    `json:"error,omitempty"`
	LatencyMs  int64     `json:"latencyMs,omitempty"`
	CheckedAt  time.Time `json:"checkedAt"`
}

// CheckProvider performs a REAL connectivity/credential check against the
// provider's own API using its merged config (Config + SecretConfig).
//
// - SMS clients that implement HealthCheckable (kavenegar) call their live API
//   (e.g. Kavenegar account/info) to validate the API key.
// - SMTP clients dial the server and complete a handshake + optional AUTH.
// - Clients with no live check implementation (unsupported types) report
//   an honest "unsupported" status instead of a fake "healthy".
func CheckProvider(ctx context.Context, p *models.Provider) *CheckResult {
	res := &CheckResult{
		ProviderID: p.ID.String(),
		Name:       p.Name,
		Channel:    p.Channel,
		Type:       p.Type,
		CheckedAt:  time.Now(),
	}

	if !p.IsEnabled || p.Status == models.ProviderStatusDisabled {
		res.Status = "disabled"
		res.Message = "Provider is disabled"
		return res
	}

	start := time.Now()
	checkErr := runCheck(ctx, p)
	res.LatencyMs = time.Since(start).Milliseconds()

	switch {
	case checkErr == nil:
		res.Status = "healthy"
		res.Message = fmt.Sprintf("%s connection verified successfully", displayName(p))
	case checkErr == errCheckUnsupported:
		res.Status = "unsupported"
		res.Message = fmt.Sprintf("Live health check is not implemented for %s provider type", p.Type)
	case errors.Is(checkErr, smsproviders.ErrNoSenderLine):
		// API key is valid but plain sends would be rejected (no sender line,
		// no senderId, no template). The provider is degraded, not down.
		res.Status = "degraded"
		res.Message = "Kavenegar API key is valid but no sender line is available: set senderId in the provider config or register/approve a sender line in the Kavenegar panel (or configure a template — template sends do not need a sender line)"
		res.Error = checkErr.Error()
	case errors.Is(checkErr, smsproviders.ErrTelegramCheckNotConfigured):
		// Telegram Gateway may be fully configured for delivery, but no
		// TELEGRAM_GATEWAY_CHECK_PHONE is set so a live probe is impossible.
		// Report degraded with an actionable message — never a fabricated
		// healthy, never a hard "down".
		res.Status = "degraded"
		res.Message = "Telegram Gateway is configured but no live health-check phone is set: configure TELEGRAM_GATEWAY_CHECK_PHONE to enable connectivity verification"
		res.Error = checkErr.Error()
	default:
		res.Status = "down"
		res.Message = fmt.Sprintf("%s health check failed", displayName(p))
		res.Error = checkErr.Error()
	}
	return res
}

var errCheckUnsupported = fmt.Errorf("health check not supported")

// runCheck builds the platform client for the provider's channel and runs its
// live Check() when the client implements the health-checkable interface.
// Unknown/unimplemented provider types are reported as "unsupported" (via
// errCheckUnsupported) rather than "down" — the config may be perfectly fine,
// the backend just has no live client for that type yet.
func runCheck(ctx context.Context, p *models.Provider) error {
	raw, err := MergeProviderConfigJSON(p)
	if err != nil {
		return err
	}

	switch p.Channel {
	case "sms":
		cfg, err := sms.ParseProviderConfig(raw)
		if err != nil {
			return fmt.Errorf("invalid sms config: %w", err)
		}
		client, err := sms.NewClientFromConfig(cfg)
		if err != nil {
			if strings.Contains(err.Error(), "unsupported or inactive SMS provider") {
				return errCheckUnsupported
			}
			return fmt.Errorf("create sms client: %w", err)
		}
		hc, ok := client.(smsproviders.HealthCheckable)
		if !ok {
			return errCheckUnsupported
		}
		return hc.Check(ctx)
	case "email":
		cfg, err := email.ParseProviderConfig(raw)
		if err != nil {
			return fmt.Errorf("invalid email config: %w", err)
		}
		client, err := email.NewClientFromConfig(cfg)
		if err != nil {
			if strings.Contains(err.Error(), "unsupported email provider") {
				return errCheckUnsupported
			}
			return fmt.Errorf("create email client: %w", err)
		}
		hc, ok := client.(email.HealthCheckable)
		if !ok {
			return errCheckUnsupported
		}
		return hc.Check(ctx)
	case "push":
		cfg, err := push.ParseProviderConfig(raw)
		if err != nil {
			return fmt.Errorf("invalid push config: %w", err)
		}
		client, err := push.NewClientFromConfig(cfg)
		if err != nil {
			if strings.Contains(err.Error(), "unsupported push provider") {
				return errCheckUnsupported
			}
			return fmt.Errorf("create push client: %w", err)
		}
		hc, ok := client.(push.HealthCheckable)
		if !ok {
			return errCheckUnsupported
		}
		return hc.Check(ctx)
	default:
		return errCheckUnsupported
	}
}

// SendTestMessage sends a REAL test message through the provider when the
// caller opts out of dry-run mode. Returns the provider-side message ID when
// the provider reports one, or a synthetic reference otherwise.
func SendTestMessage(ctx context.Context, p *models.Provider, recipient, subject, body string) (string, int64, error) {
	raw, err := MergeProviderConfigJSON(p)
	if err != nil {
		return "", 0, err
	}

	start := time.Now()
	var providerMsgID string

	switch p.Channel {
	case "sms":
		cfg, err := sms.ParseProviderConfig(raw)
		if err != nil {
			return "", 0, fmt.Errorf("invalid sms config: %w", err)
		}
		client, err := sms.NewClientFromConfig(cfg)
		if err != nil {
			return "", 0, fmt.Errorf("create sms client: %w", err)
		}
		params := map[string]string{"message": body, "body": body, "code": body, "token": body}
		if err := client.SendMessage(params, recipient); err != nil {
			return "", time.Since(start).Milliseconds(), fmt.Errorf("sms send failed: %w", err)
		}
		providerMsgID = fmt.Sprintf("test-sms-%d", time.Now().UnixNano())
	case "email":
		cfg, err := email.ParseProviderConfig(raw)
		if err != nil {
			return "", 0, fmt.Errorf("invalid email config: %w", err)
		}
		client, err := email.NewClientFromConfig(cfg)
		if err != nil {
			return "", 0, fmt.Errorf("create email client: %w", err)
		}
		if err := client.SendEmail(recipient, subject, body, false); err != nil {
			return "", time.Since(start).Milliseconds(), fmt.Errorf("email send failed: %w", err)
		}
		providerMsgID = fmt.Sprintf("test-email-%d", time.Now().UnixNano())
	case "push":
		cfg, err := push.ParseProviderConfig(raw)
		if err != nil {
			return "", 0, fmt.Errorf("invalid push config: %w", err)
		}
		client, err := push.NewClientFromConfig(cfg)
		if err != nil {
			return "", 0, fmt.Errorf("create push client: %w", err)
		}
		if err := client.SendPush(recipient, subject, body, nil); err != nil {
			return "", time.Since(start).Milliseconds(), fmt.Errorf("push send failed: %w", err)
		}
		providerMsgID = fmt.Sprintf("test-push-%d", time.Now().UnixNano())
	default:
		return "", 0, fmt.Errorf("cannot send test message for channel %q", p.Channel)
	}

	return providerMsgID, time.Since(start).Milliseconds(), nil
}

// MergeProviderConfigJSON merges a provider row's Config and SecretConfig JSON
// documents into a single JSON string (SecretConfig wins on key conflicts).
// It is the single shared implementation used by both the health checker and
// the send adapters (internal/service).
func MergeProviderConfigJSON(p *models.Provider) (string, error) {
	merged := map[string]interface{}{}
	if p.Config != "" {
		if err := json.Unmarshal([]byte(p.Config), &merged); err != nil {
			return "", fmt.Errorf("provider %s has invalid config JSON: %w", p.Name, err)
		}
	}
	if p.SecretConfig != "" {
		var secrets map[string]interface{}
		if err := json.Unmarshal([]byte(p.SecretConfig), &secrets); err != nil {
			return "", fmt.Errorf("provider %s has invalid secretConfig JSON: %w", p.Name, err)
		}
		for k, v := range secrets {
			merged[k] = v
		}
	}
	// The provider row carries its type in the dedicated Type column (e.g.
	// "kavenegar"), but the admin UI stores only channel fields (apiKey,
	// template, ...) in Config/SecretConfig — the "provider" key is omitted.
	// Client factories (sms/email/push NewClientFromConfig) switch on that key,
	// so inject the row's Type whenever the config does not already declare it.
	// This keeps the provider test endpoint, health checks, and real sends all
	// working for providers created through the UI without editing stored data.
	if _, ok := merged["provider"]; !ok && p.Type != "" {
		// Normalize to lowercase: the platform factories switch on lowercase
		// names ("kavenegar", "smtp", ...), so a stored "Kavenegar" type
		// would otherwise still hit the unsupported branch.
		merged["provider"] = strings.ToLower(p.Type)
	}
	b, err := json.Marshal(merged)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func displayName(p *models.Provider) string {
	if p.Name != "" {
		return p.Name
	}
	return p.Type
}
