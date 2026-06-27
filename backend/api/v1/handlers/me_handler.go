package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/http/middleware"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/service"
)

// MeHandler provides user-facing API endpoints.
// All endpoints derive userId from JWT claims — never from path/query/body.
type MeHandler struct {
	notificationService *service.NotificationService
	preferenceService   *service.PreferenceService
	reminderService     *service.ReminderService
}

// NewMeHandler creates a new MeHandler wrapping the underlying services.
func NewMeHandler(
	notificationService *service.NotificationService,
	preferenceService *service.PreferenceService,
	reminderService *service.ReminderService,
) *MeHandler {
	return &MeHandler{
		notificationService: notificationService,
		preferenceService:   preferenceService,
		reminderService:     reminderService,
	}
}

// getCurrentUserID extracts the user ID from the JWT claims in the Fiber context.
func (h *MeHandler) getCurrentUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
	}
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "Invalid user ID in token")
	}
	return uid, nil
}

// ============================================
// Notifications
// ============================================

// ListMyNotifications godoc
// @Summary List my notifications
// @Description Retrieve paginated notifications for the authenticated user
// @Tags User Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param channel query string false "Filter by channel"
// @Param from query string false "Start date (ISO8601)"
// @Param to query string false "End date (ISO8601)"
// @Param search query string false "Search in subject/body"
// @Success 200 {object} dto.PaginatedNotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /me/notifications [get]
func (h *MeHandler) ListMyNotifications(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)
	offset := (page - 1) * pageSize

	notifications, total, err := h.notificationService.GetUserNotifications(c.Context(), userID, pageSize, offset)
	if err != nil {
		return response.InternalError(c, "Failed to list notifications: "+err.Error())
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	result := &dto.PaginatedNotificationResponse{
		Data:       make([]*dto.NotificationResponse, 0, len(notifications)),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
	for _, notif := range notifications {
		result.Data = append(result.Data, mapNotificationToDetail(notif))
	}

	return response.OK(c, result)
}

// GetMyUnread godoc
// @Summary Get my unread notifications
// @Description Retrieve unread notifications for the authenticated user
// @Tags User Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} dto.PaginatedNotificationResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /me/notifications/unread [get]
func (h *MeHandler) GetMyUnread(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)
	offset := (page - 1) * pageSize

	notifications, total, err := h.notificationService.GetUnreadNotifications(c.Context(), userID, pageSize, offset)
	if err != nil {
		return response.InternalError(c, "Failed to list notifications: "+err.Error())
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	result := &dto.PaginatedNotificationResponse{
		Data:       make([]*dto.NotificationResponse, 0, len(notifications)),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
	for _, notif := range notifications {
		result.Data = append(result.Data, mapNotificationToDetail(notif))
	}

	return response.OK(c, result)
}

// GetMyUnreadCount godoc
// @Summary Get my unread notification count
// @Description Get the count of unread notifications for the authenticated user
// @Tags User Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UnreadCountResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /me/notifications/unread-count [get]
func (h *MeHandler) GetMyUnreadCount(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	count, err := h.notificationService.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, &dto.UnreadCountResponse{
		UserID: userID,
		Count:  count,
	})
}

// MarkMyAllRead godoc
// @Summary Mark all my notifications as read
// @Description Mark all unread notifications as read for the authenticated user
// @Tags User Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MarkAllAsReadResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /me/notifications/read-all [post]
func (h *MeHandler) MarkMyAllRead(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	count, err := h.notificationService.MarkAllAsRead(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, &dto.MarkAllAsReadResponse{
		Message:      "All notifications marked as read",
		UserID:       userID,
		UpdatedCount: count,
	})
}

// GetMyNotification godoc
// @Summary Get my notification by ID
// @Description Retrieve a single notification owned by the authenticated user
// @Tags User Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.NotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /me/notifications/{notificationId} [get]
func (h *MeHandler) GetMyNotification(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	notification, err := h.notificationService.GetNotification(c.Context(), notificationID)
	if err != nil {
		return response.NotFound(c, "Notification not found")
	}

	// Ownership check
	if notification.UserID != userID {
		return response.NotFound(c, "Notification not found")
	}

	return response.OK(c, mapNotificationToDetail(notification))
}

// MarkMyNotificationRead godoc
// @Summary Mark my notification as read
// @Description Mark a notification owned by the authenticated user as read
// @Tags User Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /me/notifications/{notificationId}/read [put]
func (h *MeHandler) MarkMyNotificationRead(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	notification, err := h.notificationService.GetNotification(c.Context(), notificationID)
	if err != nil {
		return response.NotFound(c, "Notification not found")
	}
	if notification.UserID != userID {
		return response.NotFound(c, "Notification not found")
	}

	if err := h.notificationService.MarkAsRead(c.Context(), notificationID); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, &dto.ActionResponse{
		Message: "Notification marked as read",
		ID:      notificationID,
	})
}

// MarkMyNotificationSeen godoc
// @Summary Mark my notification as seen
// @Description Mark a notification owned by the authenticated user as seen
// @Tags User Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /me/notifications/{notificationId}/seen [post]
func (h *MeHandler) MarkMyNotificationSeen(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	notification, err := h.notificationService.GetNotification(c.Context(), notificationID)
	if err != nil {
		return response.NotFound(c, "Notification not found")
	}
	if notification.UserID != userID {
		return response.NotFound(c, "Notification not found")
	}

	if err := h.notificationService.MarkAsSeen(c.Context(), notificationID); err != nil {
		return response.NotFound(c, "Notification not found")
	}

	return response.OK(c, &dto.ActionResponse{
		Message: "Notification marked as seen",
		ID:      notificationID,
	})
}

// MarkMyNotificationClicked godoc
// @Summary Mark my notification as clicked
// @Description Mark a notification owned by the authenticated user as clicked
// @Tags User Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /me/notifications/{notificationId}/click [post]
func (h *MeHandler) MarkMyNotificationClicked(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	notification, err := h.notificationService.GetNotification(c.Context(), notificationID)
	if err != nil {
		return response.NotFound(c, "Notification not found")
	}
	if notification.UserID != userID {
		return response.NotFound(c, "Notification not found")
	}

	if err := h.notificationService.MarkAsClicked(c.Context(), notificationID); err != nil {
		return response.NotFound(c, "Notification not found")
	}

	return response.OK(c, &dto.ActionResponse{
		Message: "Notification marked as clicked",
		ID:      notificationID,
	})
}

// ============================================
// Preferences
// ============================================

// GetMyPreferences godoc
// @Summary Get my notification preferences
// @Description Retrieve notification preferences for the authenticated user
// @Tags User Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []dto.PreferenceResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /me/preferences [get]
func (h *MeHandler) GetMyPreferences(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	preferences, err := h.preferenceService.GetUserPreferences(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, "Failed to retrieve preferences")
	}

	return response.OK(c, preferences)
}

// UpdateMyPreferences godoc
// @Summary Update my notification preferences
// @Description Update notification preferences for the authenticated user
// @Tags User Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param preference body dto.UpdatePreferenceRequest true "Preference data"
// @Success 200 {object} dto.PreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /me/preferences [put]
func (h *MeHandler) UpdateMyPreferences(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	req := new(dto.UpdatePreferenceRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	pref := &models.NotificationPreference{
		UserID:          userID,
		Type:            req.Type,
		IsEnabled:       req.IsEnabled,
		AllowInstant:    req.AllowInstant,
		AllowDigest:     req.AllowDigest,
		DigestFrequency: req.DigestFrequency,
	}

	if err := h.preferenceService.UpdatePreference(c.Context(), pref); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, mapPreferenceToDTO(pref))
}

// PatchMyChannelPreference godoc
// @Summary Update my channel preference
// @Description Update a specific channel preference for the authenticated user
// @Tags User Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param channel path string true "Channel (email, sms, push, in_app)"
// @Param preference body dto.ChannelPreferenceRequest true "Channel preference settings"
// @Success 200 {object} dto.PreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /me/preferences/channel/{channel} [patch]
func (h *MeHandler) PatchMyChannelPreference(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	channel := c.Params("channel")
	req := new(dto.ChannelPreferenceRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	// Load existing preference (if any)
	pref, err := h.preferenceService.GetPreferenceByUserAndType(c.Context(), userID, models.NotificationType(channel))
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	if pref == nil {
		pref = &models.NotificationPreference{
			UserID:          userID,
			Type:            models.NotificationType(channel),
			IsEnabled:       req.IsEnabled,
			AllowInstant:    true,
			AllowDigest:     true,
			DigestFrequency: "daily",
		}
	} else {
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

// PatchMyCategoryPreference godoc
// @Summary Update my category preference
// @Description Update notification category settings for the authenticated user
// @Tags User Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category path string true "Category (system, alerts, updates, marketing, security)"
// @Param preference body dto.ChannelPreferenceRequest true "Category preference settings"
// @Success 200 {object} dto.PreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /me/preferences/category/{category} [patch]
func (h *MeHandler) PatchMyCategoryPreference(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	category := c.Params("category")
	req := new(dto.ChannelPreferenceRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	preference, err := h.preferenceService.UpdateCategoryPreference(c.Context(), userID, models.NotificationCategory(category), req.IsEnabled)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, preference)
}

// ============================================
// Reminders
// ============================================

// ListMyReminders godoc
// @Summary List my reminders
// @Description Retrieve paginated reminders for the authenticated user
// @Tags User Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param scheduledFrom query string false "Start date range (ISO8601)"
// @Param scheduledTo query string false "End date range (ISO8601)"
// @Success 200 {object} dto.PaginatedResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /me/reminders [get]
func (h *MeHandler) ListMyReminders(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)
	offset := (page - 1) * pageSize

	reminders, total, err := h.reminderService.ListUserReminders(c.Context(), userID, pageSize, offset)
	if err != nil {
		return response.InternalError(c, "Failed to list reminders")
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	items := make([]*dto.ReminderResponse, 0, len(reminders))
	for _, r := range reminders {
		items = append(items, mapReminderToResponse(r))
	}

	return response.OK(c, &dto.PaginatedResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// CreateMyReminder godoc
// @Summary Create a reminder for myself
// @Description Create a new scheduled reminder for the authenticated user
// @Tags User Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-Id header string false "Tenant ID"
// @Param reminder body dto.CreateReminderRequest true "Reminder data (userId is ignored, taken from JWT)"
// @Success 201 {object} dto.ReminderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /me/reminders [post]
func (h *MeHandler) CreateMyReminder(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	req := new(dto.CreateReminderRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	reminder := &models.Reminder{
		UserID:      userID, // Force JWT userId — ignore req.UserID
		ScheduledAt: req.ScheduledAt,
		TemplateKey: req.TemplateKey,
		Subject:     req.Subject,
		Body:        req.Body,
		Status:      models.ReminderStatusPending,
	}

	// Set channels from request type
	reminder.SetChannels([]models.NotificationType{req.Type})

	// Set tenant context from header
	if tenantIDStr := c.Get("X-Tenant-Id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			reminder.TenantID = &tid
		}
	}

	// Parse recipient
	switch req.Type {
	case models.NotificationTypeEmail:
		reminder.RecipientEmail = req.Recipient
	case models.NotificationTypeSMS:
		reminder.RecipientPhone = req.Recipient
	default:
		reminder.RecipientEmail = req.Recipient
	}

	if req.Variables != nil {
		if err := reminder.SetVariables(req.Variables); err != nil {
			return response.BadRequest(c, "INVALID_VARIABLES", "Invalid variables format")
		}
	}

	if err := h.reminderService.CreateReminder(c.Context(), reminder); err != nil {
		return response.BadRequest(c, "VALIDATION_ERROR", err.Error())
	}

	return response.Created(c, mapReminderToResponse(reminder))
}

// GetMyReminder godoc
// @Summary Get my reminder by ID
// @Description Retrieve a reminder owned by the authenticated user
// @Tags User Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Success 200 {object} dto.ReminderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /me/reminders/{reminderId} [get]
func (h *MeHandler) GetMyReminder(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	reminderIDStr := c.Params("reminderId")
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_REMINDER_ID", "Invalid reminder ID")
	}

	reminder, err := h.reminderService.GetReminder(c.Context(), reminderID)
	if err != nil {
		return response.NotFound(c, "Reminder not found")
	}
	if reminder.UserID != userID {
		return response.NotFound(c, "Reminder not found")
	}

	return response.OK(c, mapReminderToResponse(reminder))
}

// UpdateMyReminder godoc
// @Summary Update my reminder
// @Description Update a scheduled reminder owned by the authenticated user
// @Tags User Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Param reminder body dto.CreateReminderRequest true "Reminder data"
// @Success 200 {object} dto.ReminderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /me/reminders/{reminderId} [put]
func (h *MeHandler) UpdateMyReminder(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	reminderIDStr := c.Params("reminderId")
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_REMINDER_ID", "Invalid reminder ID")
	}

	// Check ownership
	existing, err := h.reminderService.GetReminder(c.Context(), reminderID)
	if err != nil {
		return response.NotFound(c, "Reminder not found")
	}
	if existing.UserID != userID {
		return response.NotFound(c, "Reminder not found")
	}

	req := new(dto.CreateReminderRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	updates := &models.Reminder{
		ScheduledAt: req.ScheduledAt,
		TemplateKey: req.TemplateKey,
		Subject:     req.Subject,
		Body:        req.Body,
	}

	updates.SetChannels([]models.NotificationType{req.Type})

	switch req.Type {
	case models.NotificationTypeEmail:
		updates.RecipientEmail = req.Recipient
	case models.NotificationTypeSMS:
		updates.RecipientPhone = req.Recipient
	default:
		updates.RecipientEmail = req.Recipient
	}

	if req.Variables != nil {
		updates.SetVariables(req.Variables)
	}

	result, err := h.reminderService.UpdateReminder(c.Context(), reminderID, updates)
	if err != nil {
		if err.Error() == "reminder not found" {
			return response.NotFound(c, "Reminder not found")
		}
		return response.Conflict(c, err.Error())
	}

	return response.OK(c, mapReminderToResponse(result))
}

// CancelMyReminder godoc
// @Summary Cancel my reminder
// @Description Cancel a scheduled reminder owned by the authenticated user
// @Tags User Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Success 200 {object} dto.ReminderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /me/reminders/{reminderId}/cancel [post]
func (h *MeHandler) CancelMyReminder(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	reminderIDStr := c.Params("reminderId")
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_REMINDER_ID", "Invalid reminder ID")
	}

	// Check ownership
	existing, err := h.reminderService.GetReminder(c.Context(), reminderID)
	if err != nil {
		return response.NotFound(c, "Reminder not found")
	}
	if existing.UserID != userID {
		return response.NotFound(c, "Reminder not found")
	}

	if err := h.reminderService.CancelReminder(c.Context(), reminderID); err != nil {
		return response.Conflict(c, err.Error())
	}

	reminder, err := h.reminderService.GetReminder(c.Context(), reminderID)
	if err != nil {
		return response.NotFound(c, "Reminder not found")
	}

	return response.OK(c, mapReminderToResponse(reminder))
}

// DeleteMyReminder godoc
// @Summary Delete my reminder
// @Description Delete a scheduled or cancelled reminder owned by the authenticated user
// @Tags User Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /me/reminders/{reminderId} [delete]
func (h *MeHandler) DeleteMyReminder(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated")
	}

	reminderIDStr := c.Params("reminderId")
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_REMINDER_ID", "Invalid reminder ID")
	}

	// Check ownership
	existing, err := h.reminderService.GetReminder(c.Context(), reminderID)
	if err != nil {
		return response.NotFound(c, "Reminder not found")
	}
	if existing.UserID != userID {
		return response.NotFound(c, "Reminder not found")
	}

	if err := h.reminderService.DeleteReminder(c.Context(), reminderID); err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.OK(c, &dto.ActionResponse{
		Message: "Reminder deleted",
		ID:      reminderID,
	})
}
