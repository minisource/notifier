package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

// SettingsHandler handles admin notification settings endpoints
type SettingsHandler struct {
	settingRepo repository.SettingRepository
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(settingRepo repository.SettingRepository) *SettingsHandler {
	return &SettingsHandler{
		settingRepo: settingRepo,
	}
}

// Setting keys for notification settings
const (
	settingKeyEnabledChannels  = "notification.enabled_channels"
	settingKeyRetryPolicy      = "notification.retry_policy"
	settingKeyRateLimit        = "notification.rate_limit"
	settingKeyQuietHours       = "notification.quiet_hours"
	settingKeyRetentionDays    = "notification.retention_days"
	settingKeyDefaultProviders = "notification.default_providers"
	settingCategory            = "notification"
)

// getOrCreateSetting reads a setting by key, returning defaultValue if not found
func (h *SettingsHandler) getOrCreateSetting(ctx context.Context, key, defaultValue string) string {
	setting, err := h.settingRepo.GetByKey(ctx, key)
	if err != nil || setting == nil {
		// Create with default
		_ = h.settingRepo.Upsert(ctx, &models.Setting{
			Key:         key,
			Value:       defaultValue,
			Category:    settingCategory,
			Description: key,
			IsActive:    true,
		})
		return defaultValue
	}
	return setting.Value
}

// GetNotificationSettings godoc
// @Summary Admin: Get notification settings
// @Description Retrieve the full notification settings configuration (admin only)
// @Tags Admin Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.NotificationSettingsResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/settings/notifications [get]
func (h *SettingsHandler) GetNotificationSettings(c *fiber.Ctx) error {
	ctx := c.Context()
	settings := dto.DefaultNotificationSettings()

	// Read enabled channels
	if raw := h.getOrCreateSetting(ctx, settingKeyEnabledChannels, `{"email":true,"sms":true,"push":true,"webhook":true,"inApp":true}`); raw != "" {
		var ec dto.EnabledChannels
		if err := json.Unmarshal([]byte(raw), &ec); err == nil {
			settings.EnabledChannels = ec
		}
	}

	// Read retry policy
	if raw := h.getOrCreateSetting(ctx, settingKeyRetryPolicy, `{"enabled":true,"maxAttempts":3,"backoffStrategy":"exponential","initialDelaySeconds":60,"maxDelaySeconds":3600}`); raw != "" {
		var rp dto.RetryPolicy
		if err := json.Unmarshal([]byte(raw), &rp); err == nil {
			settings.RetryPolicy = rp
		}
	}

	// Read rate limit
	if raw := h.getOrCreateSetting(ctx, settingKeyRateLimit, `{"enabled":true,"perMinute":100,"perHour":1000}`); raw != "" {
		var rl dto.RateLimit
		if err := json.Unmarshal([]byte(raw), &rl); err == nil {
			settings.RateLimit = rl
		}
	}

	// Read quiet hours
	if raw := h.getOrCreateSetting(ctx, settingKeyQuietHours, `{}`); raw != "" && raw != "{}" {
		var qh dto.QuietHoursConfig
		if err := json.Unmarshal([]byte(raw), &qh); err == nil {
			settings.QuietHours = &qh
		}
	}

	// Read retention days
	if raw := h.getOrCreateSetting(ctx, settingKeyRetentionDays, "90"); raw != "" {
		if days, err := strconv.Atoi(raw); err == nil {
			settings.RetentionDays = days
		}
	}

	// Read default providers
	if raw := h.getOrCreateSetting(ctx, settingKeyDefaultProviders, `{}`); raw != "" && raw != "{}" {
		var dp struct {
			Email   *string `json:"email"`
			SMS     *string `json:"sms"`
			Push    *string `json:"push"`
			Webhook *string `json:"webhook"`
		}
		if err := json.Unmarshal([]byte(raw), &dp); err == nil {
			settings.DefaultEmailProviderID = dp.Email
			settings.DefaultSMSProviderID = dp.SMS
			settings.DefaultPushProviderID = dp.Push
			settings.DefaultWebhookProviderID = dp.Webhook
		}
	}

	return response.OK(c, settings)
}

// UpdateNotificationSettings godoc
// @Summary Admin: Update notification settings
// @Description Update the notification settings configuration (admin only)
// @Tags Admin Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param settings body dto.UpdateNotificationSettingsRequest true "Settings to update"
// @Success 200 {object} dto.NotificationSettingsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/settings/notifications [patch]
func (h *SettingsHandler) UpdateNotificationSettings(c *fiber.Ctx) error {
	ctx := c.Context()
	req := new(dto.UpdateNotificationSettingsRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid settings data")
	}

	// Update individual settings if provided
	if req.EnabledChannels != nil {
		data, _ := json.Marshal(req.EnabledChannels)
		if err := h.settingRepo.Upsert(ctx, &models.Setting{
			Key:         settingKeyEnabledChannels,
			Value:       string(data),
			Category:    settingCategory,
			Description: "Enabled notification channels",
			IsActive:    true,
		}); err != nil {
			return response.InternalError(c, fmt.Sprintf("Failed to update enabled channels: %v", err))
		}
	}

	if req.RetryPolicy != nil {
		data, _ := json.Marshal(req.RetryPolicy)
		if err := h.settingRepo.Upsert(ctx, &models.Setting{
			Key:         settingKeyRetryPolicy,
			Value:       string(data),
			Category:    settingCategory,
			Description: "Notification retry policy",
			IsActive:    true,
		}); err != nil {
			return response.InternalError(c, fmt.Sprintf("Failed to update retry policy: %v", err))
		}
	}

	if req.RateLimit != nil {
		data, _ := json.Marshal(req.RateLimit)
		if err := h.settingRepo.Upsert(ctx, &models.Setting{
			Key:         settingKeyRateLimit,
			Value:       string(data),
			Category:    settingCategory,
			Description: "Notification rate limits",
			IsActive:    true,
		}); err != nil {
			return response.InternalError(c, fmt.Sprintf("Failed to update rate limit: %v", err))
		}
	}

	if req.QuietHours != nil {
		data, _ := json.Marshal(req.QuietHours)
		if err := h.settingRepo.Upsert(ctx, &models.Setting{
			Key:         settingKeyQuietHours,
			Value:       string(data),
			Category:    settingCategory,
			Description: "Quiet hours configuration",
			IsActive:    true,
		}); err != nil {
			return response.InternalError(c, fmt.Sprintf("Failed to update quiet hours: %v", err))
		}
	}

	if req.RetentionDays != nil {
		val := strconv.Itoa(*req.RetentionDays)
		if err := h.settingRepo.Upsert(ctx, &models.Setting{
			Key:         settingKeyRetentionDays,
			Value:       val,
			Category:    settingCategory,
			Description: "Notification retention days",
			IsActive:    true,
		}); err != nil {
			return response.InternalError(c, fmt.Sprintf("Failed to update retention days: %v", err))
		}
	}

	// Update default providers — merge into a single JSON object
	if req.DefaultEmailProviderID != nil || req.DefaultSMSProviderID != nil ||
		req.DefaultPushProviderID != nil || req.DefaultWebhookProviderID != nil {

		// Read existing defaults
		current := `{}`
		if existing, err := h.settingRepo.GetByKey(ctx, settingKeyDefaultProviders); err == nil && existing != nil {
			current = existing.Value
		}

		var dp map[string]interface{}
		_ = json.Unmarshal([]byte(current), &dp)
		if dp == nil {
			dp = make(map[string]interface{})
		}

		if req.DefaultEmailProviderID != nil {
			dp["email"] = *req.DefaultEmailProviderID
		}
		if req.DefaultSMSProviderID != nil {
			dp["sms"] = *req.DefaultSMSProviderID
		}
		if req.DefaultPushProviderID != nil {
			dp["push"] = *req.DefaultPushProviderID
		}
		if req.DefaultWebhookProviderID != nil {
			dp["webhook"] = *req.DefaultWebhookProviderID
		}

		data, _ := json.Marshal(dp)
		if err := h.settingRepo.Upsert(ctx, &models.Setting{
			Key:         settingKeyDefaultProviders,
			Value:       string(data),
			Category:    settingCategory,
			Description: "Default providers per channel",
			IsActive:    true,
		}); err != nil {
			return response.InternalError(c, fmt.Sprintf("Failed to update default providers: %v", err))
		}
	}

	// Return the full updated settings
	return h.GetNotificationSettings(c)
}
