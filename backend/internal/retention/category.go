package retention

import "github.com/minisource/go-common/retention"

// NotifierCategory identifies a cleanup target within the Notifier service.
type NotifierCategory string

const (
	CategoryNotificationLogs          NotifierCategory = "notification_logs"
	CategoryProviderAttempts          NotifierCategory = "provider_attempts"
	CategoryProviderAttemptEvents     NotifierCategory = "provider_attempt_events"
	CategoryProviderBalanceSnapshots  NotifierCategory = "provider_balance_snapshots"
)

// CategoryMeta describes a cleanup category.
type CategoryMeta struct {
	Category         NotifierCategory
	DisplayName      string
	Description      string
	TableName        string
	MinRetentionDays int
	Protected        bool
}

// Registry returns the allowlist of cleanup categories for Notifier.
func Registry() []CategoryMeta {
	return []CategoryMeta{
		{
			Category:         CategoryNotificationLogs,
			DisplayName:      "Notification Send Logs",
			Description:      "Send lifecycle events: sending, sent, failed, retrying, held. Does NOT delete the parent notification.",
			TableName:        "notification_logs",
			MinRetentionDays: 7,
			Protected:        false,
		},
		{
			Category:         CategoryProviderAttempts,
			DisplayName:      "Provider Delivery Attempts",
			Description:      "Individual provider request/response records (SMS, email, push). Existing worker cleanup will be deprecated.",
			TableName:        "notification_provider_attempts",
			MinRetentionDays: 7,
			Protected:        false,
		},
		{
			Category:         CategoryProviderAttemptEvents,
			DisplayName:      "Provider Attempt Events",
			Description:      "Append-only event timeline for each provider attempt. Cleaned together with parent attempts.",
			TableName:        "notification_provider_attempt_events",
			MinRetentionDays: 7,
			Protected:        false,
		},
		{
			Category:         CategoryProviderBalanceSnapshots,
			DisplayName:      "Provider Balance Snapshots",
			Description:      "Historical balance/quota snapshots (already has 90-day retention). Long-term candidate.",
			TableName:        "provider_balance_snapshots",
			MinRetentionDays: 30,
			Protected:        false,
		},
		// audit_logs, notifications, delivery_control_events are NOT registered
	}
}

// IsRegistered reports whether a category is in the allowlist.
func IsRegistered(cat string) bool {
	for _, m := range Registry() {
		if string(m.Category) == cat {
			return !m.Protected
		}
	}
	return false
}

// GetMeta returns metadata for a category, or nil if not found.
func GetMeta(cat string) *CategoryMeta {
	for _, m := range Registry() {
		if string(m.Category) == cat {
			return &m
		}
	}
	return nil
}

func (c NotifierCategory) String() string { return string(c) }

const ServiceName = "notifier"

func ValidatePolicy(p *retention.Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if p.Service != ServiceName {
		return retention.ErrInvalidPolicy
	}
	meta := GetMeta(p.Category)
	if meta == nil {
		return retention.ErrCategoryProtected
	}
	if meta.Protected {
		return retention.ErrCategoryProtected
	}
	if p.Strategy == retention.StrategyAge || p.Strategy == retention.StrategyHybrid {
		if p.RetentionDays < meta.MinRetentionDays && p.CutoffTimestamp == nil {
			return retention.ErrInvalidPolicy
		}
	}
	return nil
}
