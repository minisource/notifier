package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Provider balance / quota / credit monitoring domain.
//
// A provider ACCOUNT (providers row) has one current health state
// (ProviderAccountHealth) and a durable append-only history of observations
// (ProviderBalanceSnapshot). Alert occurrences (ProviderCreditAlert) are
// deduplicated per (provider, alert_type) while active.

// Balance capability modes — what a provider account actually supports.
const (
	BalanceCapabilityAutomatic   = "automatic_balance"   // real provider account API
	BalanceCapabilityManual      = "manual_balance"      // admin-entered value
	BalanceCapabilityEstimated   = "estimated_from_usage"
	BalanceCapabilityStatusOnly  = "status_only"
	BalanceCapabilityUnsupported = "unsupported"
)

// Health levels for a provider account.
const (
	HealthLevelHealthy      = "healthy"
	HealthLevelWarning      = "warning"
	HealthLevelCritical     = "critical"
	HealthLevelExhausted    = "exhausted"
	HealthLevelStale        = "stale"
	HealthLevelUnavailable  = "unavailable" // last refresh failed / no data
	HealthLevelDisabled     = "disabled"
	HealthLevelUnsupported  = "unsupported" // provider has no balance API
)

// Alert types (per-account dedup key component).
const (
	AlertTypeWarning        = "warning"
	AlertTypeCritical       = "critical"
	AlertTypeExhausted      = "exhausted"
	AlertTypeRecovery       = "recovery"
	AlertTypeRefreshFailed  = "refresh_failed"
	AlertTypeStaleData      = "stale_data"
)

// Alert statuses.
const (
	AlertStatusActive       = "active"
	AlertStatusAcknowledged = "acknowledged"
	AlertStatusResolved     = "resolved"
)

// ProviderBalanceSnapshot is one point-in-time observation of a provider
// account's balance / quota, as reported by the provider (or entered
// manually). It is append-only. Zero and unknown are distinct: RefreshStatus
// distinguishes success/failed, and balance fields are nullable.
//
// Security: never store credentials or secret URLs. Only safe numeric/status
// values and a sanitized error message are persisted.
type ProviderBalanceSnapshot struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProviderID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"providerId"`
	TenantID      *uuid.UUID `gorm:"type:uuid;index" json:"tenantId,omitempty"`
	Provider      string     `gorm:"type:varchar(100);not null;index" json:"provider"`
	Channel       string     `gorm:"type:varchar(20)" json:"channel,omitempty"`

	// Capability + source
	CapabilityMode string `gorm:"type:varchar(50);not null;default:unsupported" json:"capabilityMode"`
	Source         string `gorm:"type:varchar(20);not null;default:provider" json:"source"` // provider | manual
	IsEstimated    bool   `gorm:"not null;default:false" json:"isEstimated"`
	IsManual       bool   `gorm:"not null;default:false" json:"isManual"`

	// Refresh outcome — NEVER conflate zero with unknown.
	RefreshStatus string `gorm:"type:varchar(20);not null;default:success" json:"refreshStatus"` // success | failed
	ErrorKind     string `gorm:"type:varchar(50)" json:"errorKind,omitempty"`                   // authentication|rate_limited|timeout|network|malformed|unknown
	ErrorCode     string `gorm:"type:varchar(100)" json:"errorCode,omitempty"`
	ErrorMessage  string `gorm:"type:text" json:"errorMessage,omitempty"` // sanitized — no credentials, no secret URLs

	// Balance / quota values (decimal via float64 — provider counts/currency
	// amounts; see docs on precision decisions in the design-system TODO).
	BalanceValue      *float64 `gorm:"type:numeric(20,4)" json:"balanceValue,omitempty"`
	BalanceUnit       string   `gorm:"type:varchar(50)" json:"balanceUnit,omitempty"` // e.g. count, rial, usd
	Currency          string   `gorm:"type:varchar(10)" json:"currency,omitempty"`
	QuotaLimit        *float64 `gorm:"type:numeric(20,4)" json:"quotaLimit,omitempty"`
	QuotaUsed         *float64 `gorm:"type:numeric(20,4)" json:"quotaUsed,omitempty"`
	QuotaRemaining    *float64 `gorm:"type:numeric(20,4)" json:"quotaRemaining,omitempty"`
	UsagePercent      *float64 `gorm:"type:numeric(8,4)" json:"usagePercent,omitempty"`
	AccountStatus     string   `gorm:"type:varchar(50)" json:"accountStatus,omitempty"`
	PlanExpiresAt     *time.Time `json:"planExpiresAt,omitempty"`

	ProviderReportedAt *time.Time `json:"providerReportedAt,omitempty"`
	FetchedAt          time.Time  `gorm:"not null;default:now();index" json:"fetchedAt"`
	LatencyMs          int64      `json:"latencyMs,omitempty"`
	RequestID          string     `gorm:"type:varchar(64)" json:"requestId,omitempty"`
	CorrelationID      string     `gorm:"type:varchar(64)" json:"correlationId,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now();index" json:"createdAt"`
}

// TableName specifies the table name.
func (ProviderBalanceSnapshot) TableName() string { return "provider_balance_snapshots" }

// BeforeCreate hook to generate UUID if not set.
func (s *ProviderBalanceSnapshot) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.CapabilityMode == "" {
		s.CapabilityMode = BalanceCapabilityUnsupported
	}
	if s.RefreshStatus == "" {
		s.RefreshStatus = "success"
	}
	if s.Source == "" {
		s.Source = "provider"
	}
	return nil
}

// ProviderAccountHealth is the current health state of one provider account.
// The last valid value is PRESERVED across refresh failures: refresh failures
// only update failure counters and HealthLevel (stale/unavailable); they never
// zero the balance.
type ProviderAccountHealth struct {
	ProviderID  uuid.UUID  `gorm:"type:uuid;primary_key" json:"providerId"`
	TenantID    *uuid.UUID `gorm:"type:uuid;index" json:"tenantId,omitempty"`
	Provider    string     `gorm:"type:varchar(100);not null" json:"provider"`
	Channel     string     `gorm:"type:varchar(20)" json:"channel,omitempty"`
	CapabilityMode string `gorm:"type:varchar(50);not null;default:unsupported" json:"capabilityMode"`

	HealthLevel string `gorm:"type:varchar(30);not null;default:unavailable" json:"healthLevel"`
	// LatestAlertLevel is the highest active alert level ("" when none).
	LatestAlertLevel string `gorm:"type:varchar(30)" json:"latestAlertLevel,omitempty"`
	AlertUpdatedAt   *time.Time `json:"alertUpdatedAt,omitempty"`

	// Latest known good values (preserved on refresh failure)
	BalanceValue   *float64 `gorm:"type:numeric(20,4)" json:"balanceValue,omitempty"`
	BalanceUnit    string   `gorm:"type:varchar(50)" json:"balanceUnit,omitempty"`
	Currency       string   `gorm:"type:varchar(10)" json:"currency,omitempty"`
	QuotaRemaining *float64 `gorm:"type:numeric(20,4)" json:"quotaRemaining,omitempty"`
	QuotaLimit     *float64 `gorm:"type:numeric(20,4)" json:"quotaLimit,omitempty"`
	UsagePercent   *float64 `gorm:"type:numeric(8,4)" json:"usagePercent,omitempty"`
	IsEstimated    bool     `gorm:"not null;default:false" json:"isEstimated"`
	IsManual       bool     `gorm:"not null;default:false" json:"isManual"`
	Source         string   `gorm:"type:varchar(20)" json:"source,omitempty"`
	LatestSnapshotID *uuid.UUID `gorm:"type:uuid" json:"latestSnapshotId,omitempty"`

	LastSuccessfulRefreshAt *time.Time `json:"lastSuccessfulRefreshAt,omitempty"`
	LastRefreshAttemptAt    *time.Time `json:"lastRefreshAttemptAt,omitempty"`
	NextScheduledRefreshAt  *time.Time `json:"nextScheduledRefreshAt,omitempty"`
	ConsecutiveFailures     int        `gorm:"not null;default:0" json:"consecutiveFailures"`
	LastErrorKind           string     `gorm:"type:varchar(50)" json:"lastErrorKind,omitempty"`
	LastErrorMessage        string     `gorm:"type:text" json:"lastErrorMessage,omitempty"`

	RefreshLockUntil *time.Time `json:"refreshLockUntil,omitempty"` // simple per-account concurrency guard

	CreatedAt time.Time `gorm:"not null;default:now()" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updatedAt"`
}

// TableName specifies the table name.
func (ProviderAccountHealth) TableName() string { return "provider_account_health" }

// BeforeSave updates the timestamp (GORM handles CreatedAt/UpdatedAt for this
// struct automatically since it has both fields).
func (h *ProviderAccountHealth) BeforeSave(tx *gorm.DB) error {
	h.UpdatedAt = time.Now().UTC()
	return nil
}

// ProviderCreditAlert is one alert occurrence. Deduplication: while an alert
// of the same (provider, alert_type) is active, no duplicate is created; the
// existing one is re-triggered (LastTriggeredAt bumped) when repeat logic
// allows. Recovery resolves the source alert and records a recovery alert.
type ProviderCreditAlert struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProviderID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"providerId"`
	TenantID      *uuid.UUID `gorm:"type:uuid;index" json:"tenantId,omitempty"`
	Provider      string     `gorm:"type:varchar(100);not null" json:"provider"`
	Channel       string     `gorm:"type:varchar(20)" json:"channel,omitempty"`

	AlertType string `gorm:"type:varchar(30);not null;index" json:"alertType"`
	Severity  string `gorm:"type:varchar(20);not null" json:"severity"` // warning|critical|exhausted|recovery|info
	Status    string `gorm:"type:varchar(20);not null;default:active;index" json:"status"`

	Message         string  `gorm:"type:text" json:"message,omitempty"` // sanitized, no secrets
	BalanceValue    *float64 `gorm:"type:numeric(20,4)" json:"balanceValue,omitempty"`
	ThresholdValue  *float64 `gorm:"type:numeric(20,4)" json:"thresholdValue,omitempty"`
	SnapshotID      *uuid.UUID `gorm:"type:uuid" json:"snapshotId,omitempty"`

	FirstTriggeredAt time.Time `gorm:"not null;default:now();index" json:"firstTriggeredAt"`
	LastTriggeredAt  time.Time `gorm:"not null;default:now()" json:"lastTriggeredAt"`
	RepeatCount      int       `gorm:"not null;default:0" json:"repeatCount"`

	AcknowledgedAt *time.Time `json:"acknowledgedAt,omitempty"`
	AcknowledgedBy string     `gorm:"type:varchar(255)" json:"acknowledgedBy,omitempty"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	ResolvedReason string     `gorm:"type:varchar(255)" json:"resolvedReason,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updatedAt"`
}

// TableName specifies the table name.
func (ProviderCreditAlert) TableName() string { return "provider_credit_alerts" }

// BeforeCreate hook to generate UUID if not set.
func (a *ProviderCreditAlert) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Status == "" {
		a.Status = AlertStatusActive
	}
	return nil
}
