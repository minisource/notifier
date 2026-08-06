package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/platform/sms"
	smsproviders "github.com/minisource/notifier/internal/platform/sms/platforms"
)

// BalanceResult is the outcome of a REAL account balance/quota fetch against a
// provider's own API. All values are provider-reported; nothing here contains
// credentials or secret URLs.
type BalanceResult struct {
	ProviderID      string     `json:"providerId"`
	Name            string     `json:"name"`
	Channel         string     `json:"channel"`
	Type            string     `json:"type,omitempty"`
	CapabilityMode  string     `json:"capabilityMode"` // automatic_balance | status_only | unsupported
	Success         bool       `json:"success"`
	BalanceValue    *float64   `json:"balanceValue,omitempty"`
	BalanceUnit     string     `json:"balanceUnit,omitempty"` // count, rial, usd ...
	Currency        string     `json:"currency,omitempty"`
	AccountStatus   string     `json:"accountStatus,omitempty"`
	PlanExpiresAt   *time.Time `json:"planExpiresAt,omitempty"`
	ProviderReportedAt *time.Time `json:"providerReportedAt,omitempty"`
	LatencyMs       int64      `json:"latencyMs,omitempty"`
	ErrorKind       string     `json:"errorKind,omitempty"` // authentication|rate_limited|timeout|network|malformed|unknown
	ErrorCode       string     `json:"errorCode,omitempty"`
	ErrorMessage    string     `json:"errorMessage,omitempty"` // sanitized
	CheckedAt       time.Time  `json:"checkedAt"`
}

// BalanceCheckable is implemented by provider clients that can report account
// balance/quota through their own API.
type BalanceCheckable interface {
	AccountInfo(ctx context.Context) (*smsproviders.KavenegarAccountInfo, error)
}

// CheckAccountBalance performs a REAL account balance fetch for a provider
// using its merged config. Providers without a balance API are reported as
// capabilityMode=unsupported (never a fake zero balance). A refresh failure
// returns Success=false with a normalized error kind — it never fabricates a
// balance value.
func CheckAccountBalance(ctx context.Context, p *models.Provider) *BalanceResult {
	res := &BalanceResult{
		ProviderID:     p.ID.String(),
		Name:           displayName(p),
		Channel:        p.Channel,
		Type:           p.Type,
		CapabilityMode: models.BalanceCapabilityUnsupported,
		CheckedAt:      time.Now(),
	}

	if !p.IsEnabled || p.Status == models.ProviderStatusDisabled {
		res.CapabilityMode = models.BalanceCapabilityUnsupported
		res.ErrorKind = "disabled"
		res.ErrorMessage = "Provider is disabled"
		return res
	}

	raw, err := MergeProviderConfigJSON(p)
	if err != nil {
		res.ErrorKind = "configuration"
		res.ErrorMessage = err.Error()
		return res
	}

	// Balance fetching is only meaningful for providers whose client exposes
	// AccountInfo (currently Kavenegar). Other SMS providers and non-SMS
	// channels are honestly marked unsupported rather than guessed.
	if p.Channel != "sms" {
		return res
	}

	cfg, err := sms.ParseProviderConfig(raw)
	if err != nil {
		res.ErrorKind = "configuration"
		res.ErrorMessage = fmt.Sprintf("invalid sms config: %v", err)
		return res
	}
	client, err := sms.NewClientFromConfig(cfg)
	if err != nil {
		res.ErrorKind = "configuration"
		res.ErrorMessage = fmt.Sprintf("create sms client: %v", err)
		return res
	}

	bc, ok := client.(BalanceCheckable)
	if !ok {
		// The client exists but has no account-info API — status-only at best.
		return res
	}

	start := time.Now()
	info, err := bc.AccountInfo(ctx)
	res.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		kind, code, msg := smsproviders.NormalizeKavenegarBalanceError(err)
		res.ErrorKind = kind
		res.ErrorCode = code
		res.ErrorMessage = msg
		return res
	}

	res.CapabilityMode = models.BalanceCapabilityAutomatic
	res.Success = true
	res.BalanceValue = info.RemainCredit
	// Kavenegar's remaincredit is a monetary credit balance in Rial (IRR), not
	// a message count. Some providers report money amounts, others report
	// message quotas — keep the unit truthful per provider.
	res.BalanceUnit = "rial"
	res.Currency = "IRR"
	res.AccountStatus = "active"
	if info.ExpireDate != nil {
		res.PlanExpiresAt = info.ExpireDate
	}
	if info.ProviderStatus != 0 {
		t := time.Now()
		res.ProviderReportedAt = &t
	}
	return res
}
