package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/service"
)

// ReminderHandler handles reminder-related endpoints
type ReminderHandler struct {
	reminderService *service.ReminderService
}

// NewReminderHandler creates a new reminder handler
func NewReminderHandler(reminderService *service.ReminderService) *ReminderHandler {
	return &ReminderHandler{reminderService: reminderService}
}

// mapReminderToResponse maps a Reminder model to a ReminderResponse DTO
func mapReminderToResponse(r *models.Reminder) *dto.ReminderResponse {
	vars := r.ParseVariables()
	chs := r.ParseChannels()
	resp := &dto.ReminderResponse{
		ID:          r.ID,
		UserID:      r.UserID,
		Recipient:   r.RecipientEmail,
		TemplateKey: r.TemplateKey,
		Locale:      "",
		Subject:     r.Subject,
		Body:        r.Body,
		Variables:   vars,
		ScheduledAt: r.ScheduledAt,
		Status:      string(r.Status),
		NotifID:     r.NotificationID,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		CancelledAt: r.CancelledAt,
	}
	if r.RecipientPhone != "" {
		resp.Recipient = r.RecipientPhone
	}
	if len(chs) > 0 {
		resp.Type = chs[0]
	}
	return resp
}

// ListReminders godoc
// @Summary List reminders
// @Description Retrieve paginated list of reminders with optional filters
// @Tags Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param userId query string false "Filter by user ID"
// @Param scheduledFrom query string false "Start of scheduled date range (ISO8601)"
// @Param scheduledTo query string false "End of scheduled date range (ISO8601)"
// @Success 200 {object} dto.PaginatedResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /reminders [get]
func (h *ReminderHandler) ListReminders(c *fiber.Ctx) error {
	filter := repository.ReminderListFilter{
		Page:          c.QueryInt("page", 1),
		PageSize:      c.QueryInt("pageSize", 20),
		SortBy:        c.Query("sortBy", "scheduled_at"),
		SortDirection: c.Query("sortDirection", "asc"),
		Search:        c.Query("search"),
	}

	if status := c.Query("status"); status != "" {
		s := models.ReminderStatus(status)
		filter.Status = &s
	}
	if userIDStr := c.Query("userId"); userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &uid
		}
	}
	if fromStr := c.Query("scheduledFrom"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.ScheduledFrom = &from
		}
	}
	if toStr := c.Query("scheduledTo"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.ScheduledTo = &to
		}
	}

	reminders, total, err := h.reminderService.ListReminders(c.Context(), filter)
	if err != nil {
		return response.InternalError(c, "Failed to list reminders: "+err.Error())
	}

	pageSize := filter.PageSize
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
		Page:       filter.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// CreateReminder godoc
// @Summary Create reminder
// @Description Create a new scheduled reminder
// @Tags Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-Id header string false "Tenant ID"
// @Param X-Request-Id header string false "Request ID"
// @Param reminder body dto.CreateReminderRequest true "Reminder data"
// @Success 201 {object} dto.ReminderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /reminders [post]
func (h *ReminderHandler) CreateReminder(c *fiber.Ctx) error {
	req := new(dto.CreateReminderRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	reminder := &models.Reminder{
		UserID:      req.UserID,
		ScheduledAt: req.ScheduledAt,
		TemplateKey: req.TemplateKey,
		Subject:     req.Subject,
		Body:        req.Body,
		Status:      models.ReminderStatusPending,
	}

	// Set tenant context
	if tenantIDStr := c.Get("X-Tenant-Id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			reminder.TenantID = &tid
		}
	}

	// Set channels from request type
	reminder.SetChannels([]models.NotificationType{req.Type})

	// Parse recipient
	switch req.Type {
	case models.NotificationTypeEmail:
		reminder.RecipientEmail = req.Recipient
	case models.NotificationTypeSMS:
		reminder.RecipientPhone = req.Recipient
	default:
		reminder.RecipientEmail = req.Recipient
	}

	// Parse variables
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

// GetReminder godoc
// @Summary Get reminder by ID
// @Description Retrieve a specific reminder
// @Tags Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Success 200 {object} dto.ReminderResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /reminders/{reminderId} [get]
func (h *ReminderHandler) GetReminder(c *fiber.Ctx) error {
	reminderIDStr := c.Params("reminderId")
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_REMINDER_ID", "Invalid reminder ID")
	}

	reminder, err := h.reminderService.GetReminder(c.Context(), reminderID)
	if err != nil {
		return response.NotFound(c, "Reminder not found")
	}

	return response.OK(c, mapReminderToResponse(reminder))
}

// UpdateReminder godoc
// @Summary Update reminder
// @Description Update an existing scheduled reminder
// @Tags Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Param reminder body dto.CreateReminderRequest true "Reminder data"
// @Success 200 {object} dto.ReminderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /reminders/{reminderId} [put]
func (h *ReminderHandler) UpdateReminder(c *fiber.Ctx) error {
	reminderIDStr := c.Params("reminderId")
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_REMINDER_ID", "Invalid reminder ID")
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
		return response.Conflict(c, err.Error())
	}

	return response.OK(c, mapReminderToResponse(result))
}

// DeleteReminder godoc
// @Summary Delete reminder
// @Description Delete a scheduled or cancelled reminder
// @Tags Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /reminders/{reminderId} [delete]
func (h *ReminderHandler) DeleteReminder(c *fiber.Ctx) error {
	reminderIDStr := c.Params("reminderId")
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_REMINDER_ID", "Invalid reminder ID")
	}

	if err := h.reminderService.DeleteReminder(c.Context(), reminderID); err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.OK(c, &dto.ActionResponse{
		Message: "Reminder deleted",
		ID:      reminderID,
	})
}

// CancelReminder godoc
// @Summary Cancel reminder
// @Description Cancel a scheduled or processing reminder
// @Tags Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reminderId path string true "Reminder ID"
// @Success 200 {object} dto.ReminderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /reminders/{reminderId}/cancel [post]
func (h *ReminderHandler) CancelReminder(c *fiber.Ctx) error {
	reminderIDStr := c.Params("reminderId")
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_REMINDER_ID", "Invalid reminder ID")
	}

	if err := h.reminderService.CancelReminder(c.Context(), reminderID); err != nil {
		return response.Conflict(c, err.Error())
	}

	// Fetch updated reminder to return full response
	reminder, err := h.reminderService.GetReminder(c.Context(), reminderID)
	if err != nil {
		return response.NotFound(c, "Reminder not found")
	}

	return response.OK(c, mapReminderToResponse(reminder))
}

// GetUserReminders godoc
// @Summary Get user reminders
// @Description Retrieve reminders for a specific user
// @Tags Reminders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /reminders/user/{userId} [get]
func (h *ReminderHandler) GetUserReminders(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)
	offset := (page - 1) * pageSize

	reminders, total, err := h.reminderService.ListUserReminders(c.Context(), userID, pageSize, offset)
	if err != nil {
		return response.InternalError(c, "Failed to list reminders: "+err.Error())
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
