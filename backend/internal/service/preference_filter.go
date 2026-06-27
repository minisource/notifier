package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

// PreferenceFilter handles preference checks before sending notifications.
// System/critical notifications (high/urgent priority, security types) can bypass preferences.
type PreferenceFilter struct {
	prefRepo repository.NotificationPreferenceRepository
	logger   logging.Logger
}

// NewPreferenceFilter creates a new preference filter
func NewPreferenceFilter(prefRepo repository.NotificationPreferenceRepository, logger logging.Logger) *PreferenceFilter {
	return &PreferenceFilter{
		prefRepo: prefRepo,
		logger:   logger,
	}
}

// FilterResult represents the result of preference filtering
type FilterResult struct {
	Allowed    bool   `json:"allowed"`
	Reason     string `json:"reason,omitempty"`
	BypassType string `json:"bypassType,omitempty"` // "system_critical", ""
}

// CheckPreference checks user preferences and determines if a notification should be sent.
// System-critical notifications (priority=urgent/high or security type) always bypass preferences.
func (f *PreferenceFilter) CheckPreference(ctx context.Context, userID uuid.UUID, notifType models.NotificationType, priority models.NotificationPriority, category string) (*FilterResult, error) {
	// System-critical bypass: urgent priority or security type always sends
	if priority == models.NotificationPriorityUrgent || notifType == models.NotificationTypeSecurity {
		f.logger.Debug(logging.General, logging.Api, "System-critical notification, bypassing preferences", map[logging.ExtraKey]interface{}{
			"userId":   userID,
			"type":     notifType,
			"priority": priority,
		})
		return &FilterResult{Allowed: true, BypassType: "system_critical"}, nil
	}

	// Get user preferences
	pref, err := f.prefRepo.GetByUserIDAndType(ctx, userID, notifType)
	if err != nil {
		// No preference found — allow with defaults
		f.logger.Debug(logging.General, logging.Api, "No preference found, using defaults (allowed)", map[logging.ExtraKey]interface{}{
			"userId": userID,
			"type":   notifType,
		})
		return &FilterResult{Allowed: true}, nil
	}

	if pref == nil {
		return &FilterResult{Allowed: true}, nil
	}

	// Check if channel is enabled at all
	if !pref.IsEnabled {
		f.logger.Info(logging.General, logging.Api, "Notification blocked: channel disabled by user preference", map[logging.ExtraKey]interface{}{
			"userId": userID,
			"type":   notifType,
		})
		return &FilterResult{Allowed: false, Reason: "channel_disabled"}, nil
	}

	// Check category settings if category is specified
	if category != "" {
		cat := models.NotificationCategory(category)
		if !pref.IsCategoryEnabled(cat) {
			f.logger.Info(logging.General, logging.Api, "Notification blocked: category disabled by user preference", map[logging.ExtraKey]interface{}{
				"userId":   userID,
				"type":     notifType,
				"category": category,
			})
			return &FilterResult{Allowed: false, Reason: "category_disabled"}, nil
		}
	}

	// Check quiet hours
	if pref.IsInQuietHours() {
		f.logger.Info(logging.General, logging.Api, "Notification deferred: user is in quiet hours", map[logging.ExtraKey]interface{}{
			"userId": userID,
			"type":   notifType,
			"quietHours": pref.QuietHours,
		})
		// In quiet hours — still allow, but mark as deferred
		return &FilterResult{Allowed: true, Reason: "quiet_hours"}, nil
	}

	// Check instant delivery preference
	if !pref.AllowInstant {
		f.logger.Info(logging.General, logging.Api, "Notification deferred: instant delivery disabled", map[logging.ExtraKey]interface{}{
			"userId": userID,
			"type":   notifType,
		})
		return &FilterResult{Allowed: true, Reason: "digest_only"}, nil
	}

	return &FilterResult{Allowed: true}, nil
}

// ParseCategoryFromMetadata extracts notification category from metadata JSON
func ParseCategoryFromMetadata(metadataJSON string) string {
	if metadataJSON == "" || metadataJSON == "{}" {
		return ""
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
		return ""
	}
	if category, ok := metadata["category"].(string); ok {
		return category
	}
	return ""
}


