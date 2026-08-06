package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/provider"
	"github.com/minisource/notifier/internal/repository"
)

// BalanceSettings are per-account thresholds. They are read from the provider's
// Config JSON under the "balanceSettings" key (non-secret), falling back to
// safe process defaults. Thresholds are expressed in the provider's own unit
// (e.g. remaining credit count for Kavenegar) — a single monetary default is
// intentionally NOT used because providers report different units.
type BalanceSettings struct {
	Enabled            bool     `json:"enabled"`                       // monitoring for this account
	WarningThreshold   *float64 `json:"warningThreshold,omitempty"`     // remaining balance below → warning
	CriticalThreshold  *float64 `json:"criticalThreshold,omitempty"`    // below → critical
	RefreshIntervalSec *int     `json:"refreshIntervalSec,omitempty"`   // override global interval
}

// DefaultBalanceSettings returns the safe defaults used when an account has no
// explicit balanceSettings in its config. Thresholds are deliberately nil
// (product decision needed per provider/unit — see design-system TODO policy
// decision table); when nil, warning/critical alerts are not triggered but
// balance is still collected, stale/unavailable and exhausted (0) states are
// still tracked.
func DefaultBalanceSettings() BalanceSettings {
	return BalanceSettings{Enabled: true}
}

// ParseBalanceSettings extracts per-account balance settings from a provider's
// merged config JSON.
func ParseBalanceSettings(configJSON string) BalanceSettings {
	s := DefaultBalanceSettings()
	if configJSON == "" {
		return s
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &raw); err != nil {
		return s
	}
	bs, ok := raw["balanceSettings"]
	if !ok {
		return s
	}
	b, err := json.Marshal(bs)
	if err != nil {
		return s
	}
	_ = json.Unmarshal(b, &s)
	return s
}

// BalanceService implements provider account balance/quota refresh, durable
// snapshot persistence, health state updates, and deduplicated credit alerts.
//
// Rules:
//   - A refresh failure NEVER replaces the last known valid balance with zero.
//   - Zero balance and unknown/unavailable balance are distinct states.
//   - Alerts are deduplicated per (provider, alert_type) while active.
//   - Recovery resolves the source alert and records a recovery alert.
//   - API keys / secret URLs are never persisted or logged (provider package
//     returns only sanitized results).
type BalanceService struct {
	cfg          *config.Config
	logger       logging.Logger
	providerRepo repository.ProviderRepository
	balanceRepo  repository.ProviderBalanceRepository
}

// NewBalanceService creates the balance service.
func NewBalanceService(
	cfg *config.Config,
	logger logging.Logger,
	providerRepo repository.ProviderRepository,
	balanceRepo repository.ProviderBalanceRepository,
) *BalanceService {
	return &BalanceService{
		cfg:          cfg,
		logger:       logger,
		providerRepo: providerRepo,
		balanceRepo:  balanceRepo,
	}
}

// RefreshAll refreshes the balance for every enabled provider account that has
// monitoring enabled. Each refresh is independent (one failure does not stop
// the rest). Used by the periodic scheduler.
func (s *BalanceService) RefreshAll(ctx context.Context) {
	providers, err := s.providerRepo.List(ctx, "", nil)
	if err != nil {
		s.logger.Warn(logging.General, logging.Select, "Balance refresh: failed to list providers", map[logging.ExtraKey]interface{}{
			logging.ExtraKey("error"): err.Error(),
		})
		return
	}
	for _, p := range providers {
		select {
		case <-ctx.Done():
			return
		default:
		}
		settings := ParseBalanceSettings(providerConfigJSON(p))
		if !settings.Enabled {
			continue
		}
		if _, err := s.RefreshAccount(ctx, p.ID); err != nil {
			s.logger.Debug(logging.General, logging.Update, "Balance refresh skipped", map[logging.ExtraKey]interface{}{
				logging.ExtraKey("providerId"): p.ID.String(),
				logging.ExtraKey("error"):      err.Error(),
			})
		}
	}
}

// providerConfigJSON returns the provider's public Config JSON, which is where
// balanceSettings (non-secret thresholds) live. SecretConfig is intentionally
// NOT included — it contains credentials and is never needed for monitoring
// settings. This differs from the send pipeline's full merge, which is why the
// helper exists separately.
func providerConfigJSON(p *models.Provider) string {
	if p == nil {
		return ""
	}
	return p.Config
}

// RefreshAccount refreshes the balance for one provider account, persists a
// snapshot, updates the health row, and evaluates alerts. It returns a
// descriptive result for API use. A failure updates failure counters and the
// health level (stale/unavailable) but preserves the last valid balance.
func (s *BalanceService) RefreshAccount(ctx context.Context, providerID uuid.UUID) (*RefreshResult, error) {
	p, err := s.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("balance refresh: get provider: %w", err)
	}
	if p == nil {
		return nil, fmt.Errorf("balance refresh: provider not found")
	}

	settings := ParseBalanceSettings(providerConfigJSON(p))

	now := time.Now()
	refreshCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.ProviderBalance.RefreshTimeoutSec)*time.Second)
	defer cancel()

	res := provider.CheckAccountBalance(refreshCtx, p)

	result := &RefreshResult{
		ProviderID:     res.ProviderID,
		Name:           res.Name,
		Channel:        res.Channel,
		CapabilityMode: res.CapabilityMode,
		Success:        res.Success,
		BalanceValue:   res.BalanceValue,
		BalanceUnit:    res.BalanceUnit,
		Currency:       res.Currency,
		ErrorKind:      res.ErrorKind,
		ErrorCode:      res.ErrorCode,
		ErrorMessage:   res.ErrorMessage,
		LatencyMs:      res.LatencyMs,
		CheckedAt:      res.CheckedAt,
	}

	// Load current health (may be nil on first refresh).
	health, _ := s.balanceRepo.GetHealth(ctx, p.ID)
	if health == nil {
		health = &models.ProviderAccountHealth{
			ProviderID:      p.ID,
			TenantID:        p.TenantID,
			Provider:        p.Type,
			Channel:         p.Channel,
			CapabilityMode:  res.CapabilityMode,
			HealthLevel:     models.HealthLevelUnavailable,
		}
	}
	health.TenantID = p.TenantID
	health.Provider = p.Type
	health.Channel = p.Channel
	health.CapabilityMode = res.CapabilityMode
	health.LastRefreshAttemptAt = &now

	// Build the snapshot row for this refresh outcome.
	snapshot := &models.ProviderBalanceSnapshot{
		ProviderID:      p.ID,
		TenantID:        p.TenantID,
		Provider:        p.Type,
		Channel:         p.Channel,
		CapabilityMode:  res.CapabilityMode,
		RefreshStatus:   "success",
		FetchedAt:       now,
		LatencyMs:       res.LatencyMs,
	}

	if res.Success {
		// Success — update latest known good values + freshness.
		health.ConsecutiveFailures = 0
		health.LastSuccessfulRefreshAt = &now
		health.LastErrorKind = ""
		health.LastErrorMessage = ""
		health.BalanceValue = res.BalanceValue
		health.BalanceUnit = res.BalanceUnit
		health.Currency = res.Currency
		health.IsEstimated = false
		health.IsManual = false
		health.Source = "provider"
		snapshot.BalanceValue = res.BalanceValue
		snapshot.BalanceUnit = res.BalanceUnit
		snapshot.Currency = res.Currency
		snapshot.AccountStatus = res.AccountStatus
		snapshot.PlanExpiresAt = res.PlanExpiresAt
		if res.ProviderReportedAt != nil {
			snapshot.ProviderReportedAt = res.ProviderReportedAt
		}
	} else {
		// Failure — preserve the last valid balance; never zero it.
		snapshot.RefreshStatus = "failed"
		snapshot.ErrorKind = res.ErrorKind
		snapshot.ErrorCode = res.ErrorCode
		snapshot.ErrorMessage = res.ErrorMessage
		health.ConsecutiveFailures++
		health.LastErrorKind = res.ErrorKind
		health.LastErrorMessage = res.ErrorMessage
	}

	// Persist snapshot (best-effort; failure is observable, not blocking).
	if err := s.balanceRepo.CreateSnapshot(ctx, snapshot); err != nil {
		s.logger.Warn(logging.General, logging.Insert, "Balance refresh: failed to persist snapshot", map[logging.ExtraKey]interface{}{
			logging.ExtraKey("providerId"): p.ID.String(),
			logging.ExtraKey("error"):      err.Error(),
		})
	}
	health.LatestSnapshotID = &snapshot.ID

	// Evaluate health level + alerts based on the refresh outcome.
	s.evaluateHealthAndAlerts(ctx, health, snapshot, settings)

	// Schedule next refresh (interval + jitter). rand.Int63n panics on 0, so
	// guard the jitter window (interval/10 can be 0 for very small intervals).
	interval := time.Duration(s.cfg.ProviderBalance.RefreshIntervalSec) * time.Second
	if settings.RefreshIntervalSec != nil && *settings.RefreshIntervalSec >= 60 {
		interval = time.Duration(*settings.RefreshIntervalSec) * time.Second
	}
	jitterWindow := interval / 10
	var jitter time.Duration
	if jitterWindow > 0 {
		jitter = time.Duration(rand.Int63n(int64(jitterWindow)))
	}
	next := now.Add(interval + jitter)
	health.NextScheduledRefreshAt = &next

	if err := s.balanceRepo.UpsertHealth(ctx, health); err != nil {
		return result, fmt.Errorf("balance refresh: upsert health: %w", err)
	}

	result.HealthLevel = health.HealthLevel
	return result, nil
}

// evaluateHealthAndAlerts computes the health level from the latest outcome and
// creates/dedupes alerts. Zero and unknown are never conflated:
//   - success with balance==0 → exhausted
//   - success below thresholds → warning/critical
//   - refresh failure → unavailable (stale if we still hold a last value older
//     than the stale window)
//   - provider unsupported → unsupported
func (s *BalanceService) evaluateHealthAndAlerts(
	ctx context.Context,
	health *models.ProviderAccountHealth,
	snapshot *models.ProviderBalanceSnapshot,
	settings BalanceSettings,
) {
	// Unsupported capability — no balance possible, not an alert condition.
	if snapshot.CapabilityMode == models.BalanceCapabilityUnsupported {
		health.HealthLevel = models.HealthLevelUnsupported
		return
	}
	if !settings.Enabled {
		health.HealthLevel = models.HealthLevelUnavailable
		return
	}

	if snapshot.RefreshStatus == "failed" {
		// A failure marks the account unavailable; when we previously had a
		// valid value and it is now older than the stale window, surface stale.
		if health.BalanceValue != nil && health.LastSuccessfulRefreshAt != nil {
			staleAfter := time.Duration(s.cfg.ProviderBalance.StaleAfterSec) * time.Second
			if time.Since(*health.LastSuccessfulRefreshAt) > staleAfter {
				health.HealthLevel = models.HealthLevelStale
			} else {
				health.HealthLevel = models.HealthLevelUnavailable
			}
		} else {
			health.HealthLevel = models.HealthLevelUnavailable
		}
		// Optional refresh-failed alert (deduped).
		s.handleRefreshFailureAlert(ctx, health)
		return
	}

	// Success path.
	balance := health.BalanceValue
	if balance == nil {
		health.HealthLevel = models.HealthLevelUnavailable
		return
	}

	// Exhausted — provider-reported zero (not unknown).
	if *balance <= 0 {
		health.HealthLevel = models.HealthLevelExhausted
		s.raiseOrDedupe(ctx, health, models.AlertTypeExhausted, models.AlertTypeExhausted, "Provider credit exhausted", *balance, 0)
		s.resolveOtherSourceAlerts(ctx, health, models.AlertTypeExhausted)
		return
	}

	// Threshold-based levels (nil threshold = not configured).
	level := models.HealthLevelHealthy
	if settings.CriticalThreshold != nil && *balance <= *settings.CriticalThreshold {
		level = models.HealthLevelCritical
		s.raiseOrDedupe(ctx, health, models.AlertTypeCritical, models.AlertTypeCritical, "Provider credit critically low", *balance, *settings.CriticalThreshold)
	} else if settings.WarningThreshold != nil && *balance <= *settings.WarningThreshold {
		level = models.HealthLevelWarning
		s.raiseOrDedupe(ctx, health, models.AlertTypeWarning, models.AlertTypeWarning, "Provider credit running low", *balance, *settings.WarningThreshold)
	} else {	// Healthy — resolve any outstanding low-credit alerts + recovery.
	s.emitRecovery(ctx, health)
	}
	health.HealthLevel = level
}

// handleRefreshFailureAlert emits a deduplicated refresh_failed alert. It never
// fabricates a balance value.
func (s *BalanceService) handleRefreshFailureAlert(ctx context.Context, health *models.ProviderAccountHealth) {
	active, err := s.balanceRepo.GetActiveAlert(ctx, health.ProviderID, models.AlertTypeRefreshFailed)
	if err != nil || active != nil {
		return // deduped (or repo error — don't spam)
	}
	msg := "Provider balance refresh failed"
	if health.LastErrorMessage != "" {
		msg = "Provider balance refresh failed: " + health.LastErrorMessage
	}
	alert := &models.ProviderCreditAlert{
		ProviderID:       health.ProviderID,
		TenantID:         health.TenantID,
		Provider:         health.Provider,
		Channel:          health.Channel,
		AlertType:        models.AlertTypeRefreshFailed,
		Severity:         "warning",
		Status:           models.AlertStatusActive,
		Message:          msg,
		BalanceValue:     health.BalanceValue, // preserved last-known value, never fabricated
		FirstTriggeredAt: time.Now(),
		LastTriggeredAt:  time.Now(),
	}
	if err := s.balanceRepo.CreateAlert(ctx, alert); err != nil {
		s.logger.Warn(logging.General, logging.Insert, "Failed to persist refresh-failed alert", map[logging.ExtraKey]interface{}{
			logging.ExtraKey("providerId"): health.ProviderID.String(),
			logging.ExtraKey("error"):      err.Error(),
		})
	}
	health.LatestAlertLevel = "warning"
	now := time.Now()
	health.AlertUpdatedAt = &now
}

// raiseOrDedupe creates an alert for the given type unless one is already
// active (dedup), in which case it re-triggers (bumps LastTriggeredAt).
func (s *BalanceService) raiseOrDedupe(ctx context.Context, health *models.ProviderAccountHealth, alertType, severity, msg string, balance, threshold float64) {
	active, err := s.balanceRepo.GetActiveAlert(ctx, health.ProviderID, alertType)
	if err == nil && active != nil {
		// Already active — dedupe. Bump the trigger time without spamming.
		active.LastTriggeredAt = time.Now()
		active.RepeatCount++
		_ = s.balanceRepo.UpdateAlert(ctx, active)
		health.LatestAlertLevel = severity
		now := time.Now()
		health.AlertUpdatedAt = &now
		return
	}
	bv := balance
	th := threshold
	alert := &models.ProviderCreditAlert{
		ProviderID:       health.ProviderID,
		TenantID:         health.TenantID,
		Provider:         health.Provider,
		Channel:          health.Channel,
		AlertType:        alertType,
		Severity:         severity,
		Status:           models.AlertStatusActive,
		Message:          msg,
		BalanceValue:     &bv,
		ThresholdValue:   &th,
		FirstTriggeredAt: time.Now(),
		LastTriggeredAt:  time.Now(),
	}
	if err := s.balanceRepo.CreateAlert(ctx, alert); err != nil {
		s.logger.Warn(logging.General, logging.Insert, "Failed to persist credit alert", map[logging.ExtraKey]interface{}{
			logging.ExtraKey("providerId"): health.ProviderID.String(),
			logging.ExtraKey("alertType"):  alertType,
			logging.ExtraKey("error"):      err.Error(),
		})
		return
	}
	health.LatestAlertLevel = severity
	now := time.Now()
	health.AlertUpdatedAt = &now
}

// resolveOtherSourceAlerts resolves warning/critical/refresh_failed alerts when
// the account becomes exhausted (exhaustion supersedes the lower severities).
func (s *BalanceService) resolveOtherSourceAlerts(ctx context.Context, health *models.ProviderAccountHealth, keepType string) {
	for _, t := range []string{models.AlertTypeWarning, models.AlertTypeCritical, models.AlertTypeRefreshFailed} {
		if t == keepType {
			continue
		}
		if _, err := s.balanceRepo.ResolveAlertsForType(ctx, health.ProviderID, t, "superseded by "+keepType); err != nil {
			s.logger.Debug(logging.General, logging.Update, "Failed to resolve superseded alert", map[logging.ExtraKey]interface{}{
				logging.ExtraKey("providerId"): health.ProviderID.String(),
				logging.ExtraKey("alertType"):  t,
			})
		}
	}
}

// emitRecovery resolves any active low-credit alerts (including refresh_failed)
// and records a recovery alert when the account returns to a healthy level.
func (s *BalanceService) emitRecovery(ctx context.Context, health *models.ProviderAccountHealth) {
	resolvedAny := false
	for _, t := range []string{models.AlertTypeWarning, models.AlertTypeCritical, models.AlertTypeExhausted, models.AlertTypeRefreshFailed} {
		if n, _ := s.balanceRepo.ResolveAlertsForType(ctx, health.ProviderID, t, "balance recovered"); n > 0 {
			resolvedAny = true
		}
	}
	health.LatestAlertLevel = ""
	health.AlertUpdatedAt = nil

	if !resolvedAny {
		return
	}
	// Recovery alert (deduped too — only when one is not already active).
	active, _ := s.balanceRepo.GetActiveAlert(ctx, health.ProviderID, models.AlertTypeRecovery)
	if active != nil {
		return
	}
	alert := &models.ProviderCreditAlert{
		ProviderID:       health.ProviderID,
		TenantID:         health.TenantID,
		Provider:         health.Provider,
		Channel:          health.Channel,
		AlertType:        models.AlertTypeRecovery,
		Severity:         "info",
		Status:           models.AlertStatusActive,
		Message:          "Provider balance recovered",
		BalanceValue:     health.BalanceValue,
		FirstTriggeredAt: time.Now(),
		LastTriggeredAt:  time.Now(),
	}
	if err := s.balanceRepo.CreateAlert(ctx, alert); err != nil {
		s.logger.Warn(logging.General, logging.Insert, "Failed to persist recovery alert", map[logging.ExtraKey]interface{}{
			logging.ExtraKey("providerId"): health.ProviderID.String(),
			logging.ExtraKey("error"):      err.Error(),
		})
	}
	// Recovery alert resolves itself after being surfaced (it is informational).
	_, _ = s.balanceRepo.ResolveAlertsForType(ctx, health.ProviderID, models.AlertTypeRecovery, "informational recovery notice")
}

// RefreshResult is the API-facing result of a manual/scheduled refresh.
type RefreshResult struct {
	ProviderID     string     `json:"providerId"`
	Name           string     `json:"name"`
	Channel        string     `json:"channel"`
	CapabilityMode string     `json:"capabilityMode"`
	Success        bool       `json:"success"`
	HealthLevel    string     `json:"healthLevel,omitempty"`
	BalanceValue   *float64   `json:"balanceValue,omitempty"`
	BalanceUnit    string     `json:"balanceUnit,omitempty"`
	Currency       string     `json:"currency,omitempty"`
	ErrorKind      string     `json:"errorKind,omitempty"`
	ErrorCode      string     `json:"errorCode,omitempty"`
	ErrorMessage   string     `json:"errorMessage,omitempty"`
	LatencyMs      int64      `json:"latencyMs,omitempty"`
	CheckedAt      time.Time  `json:"checkedAt"`
}
