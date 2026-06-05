package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

type PreferenceHandler struct {
	repo repository.NotificationPreferenceRepository
}

func NewPreferenceHandler(repo repository.NotificationPreferenceRepository) *PreferenceHandler {
	return &PreferenceHandler{repo: repo}
}

// GetUserPreferences godoc
// @Summary Get user notification preferences
// @Description Retrieve all notification preferences for a user
// @Tags Preferences
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/preferences/user/{userId} [get]
func (h *PreferenceHandler) GetUserPreferences(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
	}

	prefs, err := h.repo.GetByUserID(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, prefs)
}

// UpdatePreference godoc
// @Summary Update notification preference
// @Description Update or create notification preference for a user
// @Tags Preferences
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param preference body dto.UpdatePreferenceRequest true "Preference data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/preferences/user/{userId} [put]
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
		categoryJSON, _ := json.Marshal(req.CategorySettings)
		pref.CategorySettings = string(categoryJSON)
	} else {
		pref.CategorySettings = "{}"
	}
	if pref.QuietHours == "" {
		pref.QuietHours = "{}"
	}

	if err := h.repo.Upsert(c.Context(), pref); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, pref)
}
