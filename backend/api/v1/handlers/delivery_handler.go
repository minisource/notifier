package handlers

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/service"
)

// DeliveryHandler handles delivery-related endpoints
type DeliveryHandler struct {
	notificationService *service.NotificationService
}

// NewDeliveryHandler creates a new delivery handler
func NewDeliveryHandler(notificationService *service.NotificationService) *DeliveryHandler {
	return &DeliveryHandler{
		notificationService: notificationService,
	}
}

// notificationToDeliveryResponse converts a notification to a delivery response
func (h *DeliveryHandler) notificationToDeliveryResponse(n *models.Notification) *dto.DeliveryResponse {
	status := string(n.Status)
	if n.Status == models.NotificationStatusPending || n.Status == models.NotificationStatusQueued {
		status = "pending"
	} else if n.Status == models.NotificationStatusSending || n.Status == models.NotificationStatusProcessing {
		status = "processing"
	}

	return &dto.DeliveryResponse{
		ID:               n.ID,
		NotificationID:   n.ID,
		Provider:         n.Provider,
		Channel:          string(n.Type),
		Status:           status,
		AttemptCount:     n.RetryCount,
		MaxAttempts:      n.MaxRetries,
		LastErrorCode:    "",
		LastErrorMessage: n.ErrorMessage,
		NextRetryAt:      n.NextRetryAt,
		CreatedAt:        n.CreatedAt,
		UpdatedAt:        n.UpdatedAt,
		CompletedAt:      n.SentAt,
	}
}

// logToAttempt maps a NotificationLog to an AttemptResponse
func (h *DeliveryHandler) logToAttempt(log *models.NotificationLog, attemptNumber int) *dto.AttemptResponse {
	return &dto.AttemptResponse{
		ID:                        log.ID,
		DeliveryID:                log.NotificationID,
		AttemptNumber:             attemptNumber,
		Status:                    string(log.Status),
		ErrorMessage:              log.ErrorDetails,
		ProviderResponseSanitized: sanitizeProviderResponse(log.ProviderResponse),
		LatencyMs:                 int64(log.ProcessingTimeMs),
		CreatedAt:                 log.CreatedAt,
	}
}

// ListDeliveries godoc
// @Summary List deliveries
// @Description Retrieve paginated list of deliveries (notifications) with optional filters. Filters are applied on the notification entity which serves as the delivery unit.
// @Tags Deliveries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param status query string false "Filter by delivery status (pending, processing, sent, delivered, failed, retrying, dead, cancelled)"
// @Param channel query string false "Filter by channel (sms, email, push, in_app)"
// @Param provider query string false "Filter by provider name"
// @Param notificationId query string false "Filter by notification ID"
// @Param userId query string false "Filter by user ID"
// @Param from query string false "Start date (RFC3339)"
// @Param to query string false "End date (RFC3339)"
// @Param search query string false "Search in subject/body/recipient"
// @Param sortBy query string false "Sort field (created_at, updated_at, sent_at, priority, status)" default(created_at)
// @Param sortDirection query string false "Sort direction (asc, desc)" default(desc)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /deliveries [get]
func (h *DeliveryHandler) ListDeliveries(c *fiber.Ctx) error {
	ctx := context.Background()

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))

	filter := repository.NotificationListFilter{
		Page:          page,
		PageSize:      pageSize,
		SortBy:        c.Query("sortBy", "created_at"),
		SortDirection: c.Query("sortDirection", "desc"),
		Search:        c.Query("search"),
	}

	// Parse optional status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status := models.NotificationStatus(statusStr)
		filter.Status = &status
	}

	// Parse optional channel/type filter
	if channelStr := c.Query("channel"); channelStr != "" {
		notifType := models.NotificationType(channelStr)
		filter.Type = &notifType
	}

	// Parse optional user ID filter
	if userIDStr := c.Query("userId"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			filter.UserID = &userID
		}
	}

	// Parse optional date range
	if fromStr := c.Query("from"); fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err == nil {
			filter.From = &from
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err == nil {
			filter.To = &to
		}
	}

	notifications, total, err := h.notificationService.ListAllNotifications(ctx, filter)
	if err != nil {
		return response.InternalError(c, fmt.Sprintf("Failed to list deliveries: %s", err.Error()))
	}

	items := make([]*dto.DeliveryResponse, 0, len(notifications))
	for _, n := range notifications {
		item := h.notificationToDeliveryResponse(n)
		if n.Provider != "" && item.Provider == "" {
			item.Provider = n.Provider
		}
		item.Channel = string(n.Type)
		items = append(items, item)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	return response.OK(c, dto.PaginatedResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetDelivery godoc
// @Summary Get delivery by ID
// @Description Retrieve a specific delivery (notification) with its attempt history. Delivery ID is the same as the notification ID. Includes logs/attempts in the response.
// @Tags Deliveries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param deliveryId path string true "Delivery ID (notification UUID)"
// @Success 200 {object} dto.DeliveryResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /deliveries/{deliveryId} [get]
func (h *DeliveryHandler) GetDelivery(c *fiber.Ctx) error {
	ctx := context.Background()

	deliveryID, err := uuid.Parse(c.Params("deliveryId"))
	if err != nil {
		return response.BadRequest(c, "INVALID_DELIVERY_ID", "Invalid delivery ID format")
	}

	notification, err := h.notificationService.GetNotification(ctx, deliveryID)
	if err != nil {
		return response.NotFound(c, "Delivery not found")
	}

	delivery := h.notificationToDeliveryResponse(notification)

	// Include attempts from notification logs
	logs, err := h.notificationService.GetAttemptsFromLogs(ctx, deliveryID)
	if err == nil && len(logs) > 0 {
		attempts := make([]*dto.AttemptResponse, 0, len(logs))
		for i, log := range logs {
			attempts = append(attempts, h.logToAttempt(log, i+1))
		}
		delivery.Attempts = attempts
	}

	return response.OK(c, delivery)
}

// RetryDelivery godoc
// @Summary Retry delivery
// @Description Retry a failed/dead delivery (notification). Valid statuses for retry: failed, dead. Returns 409 Conflict if current state does not allow retry.
// @Tags Deliveries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param deliveryId path string true "Delivery ID (notification UUID)"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /deliveries/{deliveryId}/retry [post]
func (h *DeliveryHandler) RetryDelivery(c *fiber.Ctx) error {
	ctx := context.Background()

	deliveryID, err := uuid.Parse(c.Params("deliveryId"))
	if err != nil {
		return response.BadRequest(c, "INVALID_DELIVERY_ID", "Invalid delivery ID format")
	}

	err = h.notificationService.RetryNotification(ctx, deliveryID)
	if err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.OK(c, dto.ActionResponse{
		Message: "Delivery queued for retry",
		Status:  "retrying",
	})
}
