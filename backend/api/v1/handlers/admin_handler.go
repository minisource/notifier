package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/service"
)

// AdminHandler provides admin/operator API endpoints.
// All endpoints require admin or super_admin role.
// Admin can query/filter by userId, tenantId, etc.
type AdminHandler struct {
	notificationService *service.NotificationService
	templateService     *service.TemplateService
	preferenceService   *service.PreferenceService
	reminderService     *service.ReminderService
}

// NewAdminHandler creates a new AdminHandler wrapping the underlying services.
func NewAdminHandler(
	notificationService *service.NotificationService,
	templateService *service.TemplateService,
	preferenceService *service.PreferenceService,
	reminderService *service.ReminderService,
) *AdminHandler {
	return &AdminHandler{
		notificationService: notificationService,
		templateService:     templateService,
		preferenceService:   preferenceService,
		reminderService:     reminderService,
	}
}

// ============================================
// Notifications
// ============================================

// ListAllNotifications godoc
// @Summary Admin: List all notifications
// @Description Retrieve paginated list of all notifications with full filters (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param type query string false "Filter by type"
// @Param channel query string false "Filter by channel"
// @Param userId query string false "Filter by user ID"
// @Param tenantId query string false "Filter by tenant ID"
// @Param search query string false "Search in recipient/subject/body"
// @Param from query string false "Start date (ISO8601)"
// @Param to query string false "End date (ISO8601)"
// @Param sortBy query string false "Sort field" default(createdAt)
// @Param sortDirection query string false "Sort direction" default(desc)
// @Success 200 {object} dto.PaginatedNotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/notifications [get]
func (h *AdminHandler) ListAllNotifications(c *fiber.Ctx) error {
	return (&NotificationHandler{service: h.notificationService}).ListAllNotifications(c)
}

// GetNotificationByID godoc
// @Summary Admin: Get notification by ID
// @Description Retrieve a single notification with full details (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.NotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/notifications/{notificationId} [get]
func (h *AdminHandler) GetNotificationByID(c *fiber.Ctx) error {
	return (&NotificationHandler{service: h.notificationService}).GetNotificationByID(c)
}

// RetryNotification godoc
// @Summary Admin: Retry notification
// @Description Retry a failed/dead notification (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /admin/notifications/{notificationId}/retry [post]
func (h *AdminHandler) RetryNotification(c *fiber.Ctx) error {
	return (&NotificationHandler{service: h.notificationService}).RetryNotification(c)
}

// CreateNotification godoc
// @Summary Admin: Create notification
// @Description Create a new notification (admin only).\nCanonical format: {\"channel\":\"sms\",\"recipient\":{\"phone\":\"+989...\"},\"body\":\"...\"}
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notification body dto.CreateNotificationRequest true "Notification data"
// @Success 201 {object} dto.NotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/notifications [post]
func (h *AdminHandler) CreateNotification(c *fiber.Ctx) error {
	return (&NotificationHandler{service: h.notificationService}).CreateNotification(c)
}

// CancelNotification godoc
// @Summary Admin: Cancel notification
// @Description Cancel a pending/queued/retrying notification (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /admin/notifications/{notificationId}/cancel [post]
func (h *AdminHandler) CancelNotification(c *fiber.Ctx) error {
	return (&NotificationHandler{service: h.notificationService}).CancelNotification(c)
}

// GetAttempts godoc
// @Summary Admin: Get notification delivery attempts
// @Description Retrieve individual delivery attempts for a notification (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {array} dto.AttemptResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/notifications/{notificationId}/attempts [get]
func (h *AdminHandler) GetAttempts(c *fiber.Ctx) error {
	return (&NotificationHandler{service: h.notificationService}).GetNotificationAttempts(c)
}

// GetDeliveries godoc
// @Summary Admin: Get notification deliveries
// @Description Retrieve delivery attempts for a notification (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {array} dto.DeliveryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/notifications/{notificationId}/deliveries [get]
func (h *AdminHandler) GetDeliveries(c *fiber.Ctx) error {
	return (&NotificationHandler{service: h.notificationService}).GetNotificationDeliveries(c)
}

// ============================================
// Templates
// ============================================

// GetAllTemplates godoc
// @Summary Admin: List all templates
// @Description Retrieve paginated list of templates (admin only)
// @Tags Admin Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param channel query string false "Filter by channel"
// @Param locale query string false "Filter by locale"
// @Param search query string false "Search in name/key/subject"
// @Success 200 {object} dto.PaginatedResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/templates [get]
func (h *AdminHandler) GetAllTemplates(c *fiber.Ctx) error {
	return (&TemplateHandler{templateService: h.templateService}).GetAllTemplates(c)
}

// CreateTemplate godoc
// @Summary Admin: Create template
// @Description Create a new notification template (admin only)
// @Tags Admin Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param template body dto.CreateTemplateRequest true "Template data"
// @Success 201 {object} dto.TemplateResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/templates [post]
func (h *AdminHandler) CreateTemplate(c *fiber.Ctx) error {
	return (&TemplateHandler{templateService: h.templateService}).CreateTemplate(c)
}

// GetTemplateByKey godoc
// @Summary Admin: Get template by key
// @Description Retrieve a template by its unique key (admin only)
// @Tags Admin Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "Template key"
// @Success 200 {object} dto.TemplateResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/templates/key/{key} [get]
func (h *AdminHandler) GetTemplateByKey(c *fiber.Ctx) error {
	return (&TemplateHandler{templateService: h.templateService}).GetTemplateByKey(c)
}

// GetTemplate godoc
// @Summary Admin: Get template by ID
// @Description Retrieve a template by ID (admin only)
// @Tags Admin Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param templateId path string true "Template ID"
// @Success 200 {object} dto.TemplateResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/templates/{templateId} [get]
func (h *AdminHandler) GetTemplate(c *fiber.Ctx) error {
	return (&TemplateHandler{templateService: h.templateService}).GetTemplate(c)
}

// UpdateTemplate godoc
// @Summary Admin: Update template
// @Description Update an existing template (admin only)
// @Tags Admin Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param templateId path string true "Template ID"
// @Param template body dto.CreateTemplateRequest true "Template data"
// @Success 200 {object} dto.TemplateResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/templates/{templateId} [put]
func (h *AdminHandler) UpdateTemplate(c *fiber.Ctx) error {
	return (&TemplateHandler{templateService: h.templateService}).UpdateTemplate(c)
}

// DeleteTemplate godoc
// @Summary Admin: Delete template
// @Description Delete a template (admin only)
// @Tags Admin Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param templateId path string true "Template ID"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/templates/{templateId} [delete]
func (h *AdminHandler) DeleteTemplate(c *fiber.Ctx) error {
	return (&TemplateHandler{templateService: h.templateService}).DeleteTemplate(c)
}

// RenderPreviewByKey godoc
// @Summary Admin: Render preview by key
// @Description Render a template preview using template key (admin only)
// @Tags Admin Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.RenderPreviewRequest true "Render preview request"
// @Success 200 {object} dto.RenderPreviewResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/templates/render-preview [post]
func (h *AdminHandler) RenderPreviewByKey(c *fiber.Ctx) error {
	return (&TemplateHandler{templateService: h.templateService}).RenderPreviewByKey(c)
}

// RenderPreview godoc
// @Summary Admin: Render preview by ID
// @Description Render a template preview using template ID (admin only)
// @Tags Admin Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param templateId path string true "Template ID"
// @Param request body dto.RenderPreviewRequest true "Render preview request"
// @Success 200 {object} dto.RenderPreviewResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/templates/{templateId}/render-preview [post]
func (h *AdminHandler) RenderPreview(c *fiber.Ctx) error {
	return (&TemplateHandler{templateService: h.templateService}).RenderPreview(c)
}

// UpdateTemplateStatus godoc
// @Summary Admin: Update template status
// @Description Activate or deactivate a template (admin only)
// @Tags Admin Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param templateId path string true "Template ID"
// @Param status body object true "Status update {isActive: bool}"
// @Success 200 {object} dto.TemplateResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/templates/{templateId}/status [patch]
func (h *AdminHandler) UpdateTemplateStatus(c *fiber.Ctx) error {
	return (&TemplateHandler{templateService: h.templateService}).UpdateTemplateStatus(c)
}

// ============================================
// Preferences (admin can manage any user's preferences)
// ============================================

// GetUserPreferences godoc
// @Summary Admin: Get user preferences
// @Description Retrieve notification preferences for any user (admin only)
// @Tags Admin Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Success 200 {object} []dto.PreferenceResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/preferences/user/{userId} [get]
func (h *AdminHandler) GetUserPreferences(c *fiber.Ctx) error {
	return (&PreferenceHandler{preferenceService: h.preferenceService}).GetUserPreferences(c)
}

// UpdatePreference godoc
// @Summary Admin: Update user preferences
// @Description Update notification preferences for any user (admin only)
// @Tags Admin Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param preference body dto.UpdatePreferenceRequest true "Preference data"
// @Success 200 {object} dto.PreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/preferences/user/{userId} [put]
func (h *AdminHandler) UpdatePreference(c *fiber.Ctx) error {
	return (&PreferenceHandler{preferenceService: h.preferenceService}).UpdatePreference(c)
}

// UpdateChannelPreference godoc
// @Summary Admin: Update user channel preference
// @Description Update a specific channel preference for any user (admin only)
// @Tags Admin Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param channel path string true "Channel"
// @Param preference body dto.ChannelPreferenceRequest true "Channel preference settings"
// @Success 200 {object} dto.PreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/preferences/user/{userId}/channel/{channel} [patch]
func (h *AdminHandler) UpdateChannelPreference(c *fiber.Ctx) error {
	return (&PreferenceHandler{preferenceService: h.preferenceService}).UpdateChannelPreference(c)
}

// UpdateCategoryPreference godoc
// @Summary Admin: Update user category preference
// @Description Update notification category settings for any user (admin only)
// @Tags Admin Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param category path string true "Category"
// @Param preference body dto.ChannelPreferenceRequest true "Category preference settings"
// @Success 200 {object} dto.PreferenceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/preferences/user/{userId}/category/{category} [patch]
func (h *AdminHandler) UpdateCategoryPreference(c *fiber.Ctx) error {
	return (&PreferenceHandler{preferenceService: h.preferenceService}).UpdateCategoryPreference(c)
}

// ============================================
// Reminders
// ============================================

// ListAllReminders godoc
// @Summary Admin: List all reminders
// @Description Retrieve paginated list of all reminders with filters (admin only)
// @Tags Admin Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param userId query string false "Filter by user ID"
// @Param scheduledFrom query string false "Start date range (ISO8601)"
// @Param scheduledTo query string false "End date range (ISO8601)"
// @Param search query string false "Search in subject/templateKey"
// @Success 200 {object} dto.PaginatedResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/reminders [get]
func (h *AdminHandler) ListAllReminders(c *fiber.Ctx) error {
	return (&ReminderHandler{reminderService: h.reminderService}).ListReminders(c)
}

// CreateReminder godoc
// @Summary Admin: Create reminder
// @Description Create a scheduled reminder for any user (admin only)
// @Tags Admin Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-Id header string false "Tenant ID"
// @Param reminder body dto.CreateReminderRequest true "Reminder data"
// @Success 201 {object} dto.ReminderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/reminders [post]
func (h *AdminHandler) CreateReminder(c *fiber.Ctx) error {
	return (&ReminderHandler{reminderService: h.reminderService}).CreateReminder(c)
}

// GetReminder godoc
// @Summary Admin: Get reminder by ID
// @Description Retrieve a specific reminder (admin only)
// @Tags Admin Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Success 200 {object} dto.ReminderResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/reminders/{reminderId} [get]
func (h *AdminHandler) GetReminder(c *fiber.Ctx) error {
	return (&ReminderHandler{reminderService: h.reminderService}).GetReminder(c)
}

// UpdateReminder godoc
// @Summary Admin: Update reminder
// @Description Update a scheduled reminder for any user (admin only)
// @Tags Admin Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Param reminder body dto.CreateReminderRequest true "Reminder data"
// @Success 200 {object} dto.ReminderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /admin/reminders/{reminderId} [put]
func (h *AdminHandler) UpdateReminder(c *fiber.Ctx) error {
	return (&ReminderHandler{reminderService: h.reminderService}).UpdateReminder(c)
}

// CancelReminder godoc
// @Summary Admin: Cancel reminder
// @Description Cancel a scheduled/processing reminder for any user (admin only)
// @Tags Admin Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Success 200 {object} dto.ReminderResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /admin/reminders/{reminderId}/cancel [post]
func (h *AdminHandler) CancelReminder(c *fiber.Ctx) error {
	return (&ReminderHandler{reminderService: h.reminderService}).CancelReminder(c)
}

// DeleteReminder godoc
// @Summary Admin: Delete reminder
// @Description Delete a scheduled/cancelled reminder (admin only)
// @Tags Admin Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /admin/reminders/{reminderId} [delete]
func (h *AdminHandler) DeleteReminder(c *fiber.Ctx) error {
	return (&ReminderHandler{reminderService: h.reminderService}).DeleteReminder(c)
}

// MarkAsSeen godoc
// @Summary Admin: Mark notification as seen
// @Description Mark a notification as seen (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/notifications/{notificationId}/seen [post]
func (h *AdminHandler) MarkAsSeen(c *fiber.Ctx) error {
	return (&NotificationHandler{service: h.notificationService}).MarkAsSeen(c)
}

// MarkAsClicked godoc
// @Summary Admin: Mark notification as clicked
// @Description Mark a notification as clicked (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/notifications/{notificationId}/click [post]
func (h *AdminHandler) MarkAsClicked(c *fiber.Ctx) error {
	return (&NotificationHandler{service: h.notificationService}).MarkAsClicked(c)
}

// MarkAsRead godoc
// @Summary Admin: Mark notification as read
// @Description Mark a notification as read (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/notifications/{notificationId}/read [put]
func (h *AdminHandler) MarkAsRead(c *fiber.Ctx) error {
	return (&NotificationHandler{service: h.notificationService}).MarkAsRead(c)
}

// ReadAllNotifications godoc
// @Summary Admin: Mark all notifications as read
// @Description Mark all unread notifications as read for a user (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId query string true "User ID"
// @Success 200 {object} dto.MarkAllAsReadResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/notifications/read-all [post]
func (h *AdminHandler) ReadAllNotifications(c *fiber.Ctx) error {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		return response.BadRequest(c, "USER_ID_REQUIRED", "userId query parameter is required")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
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

// GetUserReminders godoc
// @Summary Admin: Get user reminders
// @Description Retrieve reminders for a specific user (admin only)
// @Tags Admin Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/reminders/user/{userId} [get]
func (h *AdminHandler) GetUserReminders(c *fiber.Ctx) error {
	return (&ReminderHandler{reminderService: h.reminderService}).GetUserReminders(c)
}
