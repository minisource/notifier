package handlers

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/service"
)

// ProviderAttemptHandler serves the durable Provider Request Lifecycle Logging
// API (admin-only). All payloads returned are already sanitized server-side.
type ProviderAttemptHandler struct {
	notificationService *service.NotificationService
}

// NewProviderAttemptHandler creates the handler.
func NewProviderAttemptHandler(notificationService *service.NotificationService) *ProviderAttemptHandler {
	return &ProviderAttemptHandler{notificationService: notificationService}
}

// ListProviderAttempts godoc
// @Summary List provider attempts
// @Description Paginated provider request lifecycle records with filters (admin only)
// @Tags Provider Attempts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size (max 100)" default(20)
// @Param notificationId query string false "Filter by notification ID"
// @Param channel query string false "Filter by channel (sms, email, push, in_app)"
// @Param provider query string false "Filter by provider name"
// @Param status query string false "Filter by lifecycle status (queued, preparing, sending, accepted, pending, delivered, failed, rejected, timed_out, cancelled, unknown)"
// @Param providerMessageId query string false "Filter by provider message ID"
// @Param requestId query string false "Filter by request ID"
// @Param correlationId query string false "Filter by correlation ID"
// @Param from query string false "Start date (RFC3339)"
// @Param to query string false "End date (RFC3339)"
// @Success 200 {object} dto.ProviderAttemptListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/attempts [get]
func (h *ProviderAttemptHandler) ListProviderAttempts(c *fiber.Ctx) error {
	ctx := context.Background()
	repo := h.notificationService.GetProviderAttemptRepository()
	if repo == nil {
		return response.InternalError(c, "Provider attempt logging is not available")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))

	filter := repository.ProviderAttemptFilter{
		Page:              page,
		PageSize:          pageSize,
		Channel:           c.Query("channel"),
		Provider:          c.Query("provider"),
		Status:            c.Query("status"),
		ProviderMessageID: c.Query("providerMessageId"),
		RequestID:         c.Query("requestId"),
		CorrelationID:     c.Query("correlationId"),
	}

	if nid := c.Query("notificationId"); nid != "" {
		if id, err := uuid.Parse(nid); err == nil {
			filter.NotificationID = &id
		}
	}
	if tid := resolveTenantID(c); tid != nil {
		filter.TenantID = tid
	}

	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.From = &from
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.To = &to
		}
	}

	attempts, total, err := repo.List(ctx, filter)
	if err != nil {
		return response.InternalError(c, "Failed to list provider attempts: "+err.Error())
	}

	items := make([]*dto.ProviderAttemptSummary, 0, len(attempts))
	for _, a := range attempts {
		items = append(items, dto.MapProviderAttemptSummary(a))
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.PageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	return response.OK(c, dto.ProviderAttemptListResponse{
		Items:      items,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	})
}

// GetProviderAttempt godoc
// @Summary Get provider attempt details
// @Description Retrieve a single provider attempt with sanitized request/response and lifecycle events (admin only)
// @Tags Provider Attempts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param attemptId path string true "Provider attempt ID"
// @Success 200 {object} dto.ProviderAttemptDetails
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/attempts/{attemptId} [get]
func (h *ProviderAttemptHandler) GetProviderAttempt(c *fiber.Ctx) error {
	ctx := context.Background()
	repo := h.notificationService.GetProviderAttemptRepository()
	if repo == nil {
		return response.InternalError(c, "Provider attempt logging is not available")
	}

	attemptID, err := uuid.Parse(c.Params("attemptId"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ATTEMPT_ID", "Invalid provider attempt ID")
	}

	attempt, err := repo.GetAttemptByID(ctx, attemptID)
	if err != nil {
		return response.NotFound(c, "Provider attempt not found")
	}

	// Tenant isolation
	if tid := resolveTenantID(c); tid != nil && attempt.TenantID != nil && *attempt.TenantID != *tid {
		return response.NotFound(c, "Provider attempt not found")
	}

	events, evErr := repo.ListEventsByAttempt(ctx, attemptID)
	if evErr != nil {
		events = nil
	}

	return response.OK(c, dto.MapProviderAttemptDetails(attempt, events))
}

// ListProviderAttemptEvents godoc
// @Summary List provider attempt events
// @Description Retrieve the lifecycle timeline for a provider attempt (admin only)
// @Tags Provider Attempts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param attemptId path string true "Provider attempt ID"
// @Success 200 {array} dto.ProviderAttemptEventResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/attempts/{attemptId}/events [get]
func (h *ProviderAttemptHandler) ListProviderAttemptEvents(c *fiber.Ctx) error {
	ctx := context.Background()
	repo := h.notificationService.GetProviderAttemptRepository()
	if repo == nil {
		return response.InternalError(c, "Provider attempt logging is not available")
	}

	attemptID, err := uuid.Parse(c.Params("attemptId"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ATTEMPT_ID", "Invalid provider attempt ID")
	}

	attempt, err := repo.GetAttemptByID(ctx, attemptID)
	if err != nil {
		return response.NotFound(c, "Provider attempt not found")
	}
	if tid := resolveTenantID(c); tid != nil && attempt.TenantID != nil && *attempt.TenantID != *tid {
		return response.NotFound(c, "Provider attempt not found")
	}

	events, err := repo.ListEventsByAttempt(ctx, attemptID)
	if err != nil {
		return response.InternalError(c, "Failed to list attempt events: "+err.Error())
	}

	out := make([]*dto.ProviderAttemptEventResponse, 0, len(events))
	for _, e := range events {
		out = append(out, dto.MapProviderAttemptEvent(e))
	}
	return response.OK(c, out)
}

// ListNotificationAttempts godoc
// @Summary List attempts for a notification
// @Description Retrieve the provider attempt history for a single notification (admin only)
// @Tags Provider Attempts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {array} dto.ProviderAttemptSummary
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/notifications/{notificationId}/attempts [get]
func (h *ProviderAttemptHandler) ListNotificationAttempts(c *fiber.Ctx) error {
	ctx := context.Background()
	repo := h.notificationService.GetProviderAttemptRepository()
	if repo == nil {
		return response.InternalError(c, "Provider attempt logging is not available")
	}

	notificationID, err := uuid.Parse(c.Params("notificationId"))
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	attempts, err := repo.ListByNotification(ctx, notificationID)
	if err != nil {
		return response.InternalError(c, "Failed to list attempts: "+err.Error())
	}

	items := make([]*dto.ProviderAttemptSummary, 0, len(attempts))
	for _, a := range attempts {
		items = append(items, dto.MapProviderAttemptSummary(a))
	}
	return response.OK(c, items)
}
