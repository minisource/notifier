package dto

import (
	"time"

	"github.com/google/uuid"
)

// ProviderAccountHealthSummary is the dashboard-facing health of one provider
// account. Explicit availability states (stale/unavailable/unsupported/manual/
// estimated) are surfaced; unknown values are null, never zero.
type ProviderAccountHealthSummary struct {
	ProviderID            string     `json:"providerId"`
	TenantID              *uuid.UUID `json:"tenantId,omitempty"`
	Provider              string     `json:"provider"`
	Channel               string     `json:"channel,omitempty"`
	CapabilityMode        string     `json:"capabilityMode"`
	HealthLevel           string     `json:"healthLevel"`
	LatestAlertLevel      string     `json:"latestAlertLevel,omitempty"`
	BalanceValue          *float64   `json:"balanceValue,omitempty"`
	BalanceUnit           string     `json:"balanceUnit,omitempty"`
	Currency              string     `json:"currency,omitempty"`
	QuotaRemaining        *float64   `json:"quotaRemaining,omitempty"`
	QuotaLimit            *float64   `json:"quotaLimit,omitempty"`
	UsagePercent          *float64   `json:"usagePercent,omitempty"`
	IsEstimated           bool       `json:"isEstimated"`
	IsManual              bool       `json:"isManual"`
	Source                string     `json:"source,omitempty"`
	LastSuccessfulRefreshAt *time.Time `json:"lastSuccessfulRefreshAt,omitempty"`
	LastRefreshAttemptAt    *time.Time `json:"lastRefreshAttemptAt,omitempty"`
	NextScheduledRefreshAt  *time.Time `json:"nextScheduledRefreshAt,omitempty"`
	ConsecutiveFailures     int        `json:"consecutiveFailures"`
	LastErrorKind           string     `json:"lastErrorKind,omitempty"`
	LastErrorMessage        string     `json:"lastErrorMessage,omitempty"`
	ActiveAlertCount        int        `json:"activeAlertCount"`
	UpdatedAt               time.Time  `json:"updatedAt"`
}

// BalanceSnapshotDTO is one point-in-time observation.
type BalanceSnapshotDTO struct {
	ID              string     `json:"id"`
	ProviderID      string     `json:"providerId"`
	RefreshStatus   string     `json:"refreshStatus"` // success | failed
	CapabilityMode  string     `json:"capabilityMode"`
	Source          string     `json:"source"`
	IsEstimated     bool       `json:"isEstimated"`
	IsManual        bool       `json:"isManual"`
	BalanceValue    *float64   `json:"balanceValue,omitempty"`
	BalanceUnit     string     `json:"balanceUnit,omitempty"`
	Currency        string     `json:"currency,omitempty"`
	QuotaRemaining  *float64   `json:"quotaRemaining,omitempty"`
	QuotaLimit      *float64   `json:"quotaLimit,omitempty"`
	UsagePercent    *float64   `json:"usagePercent,omitempty"`
	AccountStatus   string     `json:"accountStatus,omitempty"`
	PlanExpiresAt   *time.Time `json:"planExpiresAt,omitempty"`
	ErrorKind       string     `json:"errorKind,omitempty"`
	ErrorCode       string     `json:"errorCode,omitempty"`
	ErrorMessage    string     `json:"errorMessage,omitempty"`
	FetchedAt       time.Time  `json:"fetchedAt"`
	LatencyMs       int64      `json:"latencyMs,omitempty"`
}

// CreditAlertDTO is an alert occurrence.
type CreditAlertDTO struct {
	ID               string     `json:"id"`
	ProviderID       string     `json:"providerId"`
	Provider         string     `json:"provider"`
	AlertType        string     `json:"alertType"`
	Severity         string     `json:"severity"`
	Status           string     `json:"status"`
	Message          string     `json:"message,omitempty"`
	BalanceValue     *float64   `json:"balanceValue,omitempty"`
	ThresholdValue   *float64   `json:"thresholdValue,omitempty"`
	FirstTriggeredAt time.Time  `json:"firstTriggeredAt"`
	LastTriggeredAt  time.Time  `json:"lastTriggeredAt"`
	RepeatCount      int        `json:"repeatCount"`
	AcknowledgedAt   *time.Time `json:"acknowledgedAt,omitempty"`
	AcknowledgedBy   string     `json:"acknowledgedBy,omitempty"`
	ResolvedAt       *time.Time `json:"resolvedAt,omitempty"`
	ResolvedReason   string     `json:"resolvedReason,omitempty"`
}

// BalanceSettingsDTO is the per-account monitoring configuration.
type BalanceSettingsDTO struct {
	Enabled            bool     `json:"enabled"`
	WarningThreshold   *float64 `json:"warningThreshold,omitempty"`
	CriticalThreshold  *float64 `json:"criticalThreshold,omitempty"`
	RefreshIntervalSec *int     `json:"refreshIntervalSec,omitempty"`
}

// BalanceRefreshResponse is the result of a manual refresh.
type BalanceRefreshResponse struct {
	ProviderID     string    `json:"providerId"`
	Name           string    `json:"name"`
	Channel        string    `json:"channel"`
	CapabilityMode string    `json:"capabilityMode"`
	Success        bool      `json:"success"`
	HealthLevel    string    `json:"healthLevel,omitempty"`
	BalanceValue   *float64  `json:"balanceValue,omitempty"`
	BalanceUnit    string    `json:"balanceUnit,omitempty"`
	Currency       string    `json:"currency,omitempty"`
	ErrorKind      string    `json:"errorKind,omitempty"`
	ErrorCode      string    `json:"errorCode,omitempty"`
	ErrorMessage   string    `json:"errorMessage,omitempty"`
	LatencyMs      int64     `json:"latencyMs,omitempty"`
	CheckedAt      time.Time `json:"checkedAt"`
}
