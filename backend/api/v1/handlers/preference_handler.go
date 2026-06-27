package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/service"
)

type PreferenceHandler struct {
	preferenceService *service.PreferenceService
}

func NewPreferenceHandler(preferenceService *service.PreferenceService) *PreferenceHandler {
	return &PreferenceHandler{preferenceService: preferenceService}
}

// GetUserPreferences godoc
// @Summary Get user notification preferences
// @Description Retrieve all notification preferences for a user
// @Tags Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param X-Tenant-Id header string false "Tenant ID"
// @Param X-Request-Id header string false "Request ID"
// @Success 200 {array} dto.PreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /preferences/user/{userId} [get]
func (h *PreferenceHandler) GetUserPreferences(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
	}

	prefs, err := h.preferenceService.GetUserPreferences(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	// Map to response DTOs
	responses := make([]dto.PreferenceResponse, 0, len(prefs))
	for _, p := range prefs {
		responses = append(responses, mapPreferenceToDTO(p))
	}

	return response.OK(c, responses)
}

// UpdatePreference godoc
// @Summary Update notification preference
// @Description Update or create notification preference for a user
// @Tags Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param X-Tenant-Id header string false "Tenant ID"
// @Param X-Request-Id header string false "Request ID"
// @Param preference body dto.UpdatePreferenceRequest true "Preference data"
// @Success 200 {object} dto.PreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /preferences/user/{userId} [put]
func (h *PreferenceHandler) UpdatePreference(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
	}

	req := new(dto.UpdatePreferenceRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", i18n.T(c.Context(), "errors.invalid_request"))
	}

	pref := &models.NotificationPreference{
		UserID:          userID,
		Type:            req.Type,
		IsEnabled:       req.IsEnabled,
		AllowInstant:    req.AllowInstant,
		AllowDigest:     req.AllowDigest,
		DigestFrequency: req.DigestFrequency,
	}

	if req.CategorySettings != nil {
		if err := pref.SetCategorySettings(req.CategorySettings); err != nil {
			return response.BadRequest(c, "INVALID_CATEGORY_SETTINGS", "Invalid category settings format")
		}
	}
	if req.QuietHours != nil {
		if err := pref.SetQuietHours(req.QuietHours); err != nil {
			return response.BadRequest(c, "INVALID_QUIET_HOURS", "Invalid quiet hours format")
		}
	}
	if pref.QuietHours == "" {
		pref.QuietHours = "{}"
	}

	if err := h.preferenceService.UpdatePreference(c.Context(), pref); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, mapPreferenceToDTO(pref))
}

// UpdateChannelPreference godoc
// @Summary Update channel-specific preference
// @Description Update preferences for a specific notification channel (sms, email, push, in_app)
// @Tags Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param channel path string true "Notification channel (sms, email, push, in_app)"
// @Param X-Tenant-Id header string false "Tenant ID"
// @Param X-Request-Id header string false "Request ID"
// @Param preference body dto.ChannelPreferenceRequest true "Channel preference data"
// @Success 200 {object} dto.PreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /preferences/user/{userId}/channel/{channel} [patch]
func (h *PreferenceHandler) UpdateChannelPreference(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
	}

	channel := c.Params("channel")
	// Validate channel
	switch models.NotificationType(channel) {
	case models.NotificationTypeSMS, models.NotificationTypeEmail, models.NotificationTypePush, models.NotificationTypeInApp:
		// Valid
	default:
		return response.BadRequest(c, "INVALID_CHANNEL", "Invalid notification channel. Use: sms, email, push, in_app")
	}

	req := new(dto.ChannelPreferenceRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", i18n.T(c.Context(), "errors.invalid_request"))
	}

	// Load existing preferences to preserve unset fields (PATCH semantics)
	pref, err := h.preferenceService.GetPreferenceByUserAndType(c.Context(), userID, models.NotificationType(channel))
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	if pref == nil {
		// No existing preference — start with defaults
		pref = &models.NotificationPreference{
			UserID:          userID,
			Type:            models.NotificationType(channel),
			IsEnabled:       req.IsEnabled,
			AllowInstant:    true,
			AllowDigest:     true,
			DigestFrequency: "daily",
			QuietHours:      "{}",
		}
	} else {
		// Only update fields explicitly provided in the request
		pref.IsEnabled = req.IsEnabled
		if req.AllowInstant != nil {
			pref.AllowInstant = *req.AllowInstant
		}
		if req.AllowDigest != nil {
			pref.AllowDigest = *req.AllowDigest
		}
		if req.DigestFrequency != "" {
			pref.DigestFrequency = req.DigestFrequency
		}
		if req.QuietHours != nil {
			if err := pref.SetQuietHours(req.QuietHours); err != nil {
				return response.BadRequest(c, "INVALID_QUIET_HOURS", "Invalid quiet hours format")
			}
		}
	}

	if err := h.preferenceService.UpdatePreference(c.Context(), pref); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, mapPreferenceToDTO(pref))
}

// UpdateCategoryPreference godoc
// @Summary Update category preference
// @Description Update notification preferences for a specific category
// @Tags Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param category path string true "Notification category (system, alerts, updates, marketing, security)"
// @Param X-Tenant-Id header string false "Tenant ID"
// @Param X-Request-Id header string false "Request ID"
// @Param preference body dto.ChannelPreferenceRequest true "Category preference data"
// @Success 200 {object} dto.PreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /preferences/user/{userId}/category/{category} [patch]
func (h *PreferenceHandler) UpdateCategoryPreference(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
	}

	category := c.Params("category")
	// Validate category
	switch models.NotificationCategory(category) {
	case models.NotificationCategorySystem, models.NotificationCategoryAlerts, models.NotificationCategoryUpdates, models.NotificationCategoryMarketing, models.NotificationCategorySecurity:
		// Valid
	default:
		return response.BadRequest(c, "INVALID_CATEGORY", "Invalid notification category. Use: system, alerts, updates, marketing, security")
	}

	req := new(dto.ChannelPreferenceRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", i18n.T(c.Context(), "errors.invalid_request"))
	}

	// Load existing preferences for this user and update category settings
	prefs, err := h.preferenceService.GetUserPreferences(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	// Update category setting across all preference channels
	for _, pref := range prefs {
		cs := pref.ParseCategorySettings()
		if cs == nil {
			cs = make(map[models.NotificationCategory]bool)
		}
		cs[models.NotificationCategory(category)] = req.IsEnabled
		if err := pref.SetCategorySettings(cs); err != nil {
			return response.BadRequest(c, "INVALID_CATEGORY_SETTINGS", "Invalid category settings format")
		}
		if err := h.preferenceService.UpdatePreference(c.Context(), pref); err != nil {
			return response.InternalError(c, err.Error())
		}
	}

	return response.OK(c, fiber.Map{
		"message":  "Category preference updated",
		"userId":   userID,
		"category": category,
	})
}

// mapPreferenceToDTO maps a NotificationPreference model to a PreferenceResponse DTO
func mapPreferenceToDTO(p *models.NotificationPreference) dto.PreferenceResponse {
	resp := dto.PreferenceResponse{
		ID:              p.ID,
		UserID:          p.UserID,
		Type:            p.Type,
		IsEnabled:       p.IsEnabled,
		AllowInstant:    p.AllowInstant,
		AllowDigest:     p.AllowDigest,
		DigestFrequency: p.DigestFrequency,
	}

	// Parse quiet hours
	if qh, err := p.ParseQuietHours(); err == nil && qh != nil && qh.Start != "" {
		resp.QuietHours = qh
	}

	// Parse category settings
	if cs := p.ParseCategorySettings(); cs != nil {
		resp.CategorySettings = cs
	}

	return resp
}
