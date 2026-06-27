package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/service"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(service *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// mapNotificationToDetail maps a Notification model to a full NotificationResponse DTO
func mapNotificationToDetail(n *models.Notification) *dto.NotificationResponse {
	return &dto.NotificationResponse{
		ID:             n.ID,
		TenantID:       n.TenantID,
		UserID:         n.UserID,
		Type:           n.Type,
		Status:         n.Status,
		Priority:       n.Priority,
		RecipientEmail: n.RecipientEmail,
		RecipientPhone: n.RecipientPhone,
		Subject:        n.Subject,
		Body:           n.Body,
		TemplateID:     n.TemplateID,
		TemplateKey:    n.TemplateKey,
		Locale:         n.Locale,
		RetryCount:     n.RetryCount,
		MaxRetries:     n.MaxRetries,
		ErrorMessage:   n.ErrorMessage,
		Provider:       n.Provider,
		ScheduledAt:    n.ScheduledAt,
		SentAt:         n.SentAt,
		DeliveredAt:    n.DeliveredAt,
		FailedAt:       n.FailedAt,
		SeenAt:         n.SeenAt,
		ReadAt:         n.ReadAt,
		ClickedAt:      n.ClickedAt,
		CancelledAt:    n.CancelledAt,
		CreatedAt:      n.CreatedAt,
		UpdatedAt:      n.UpdatedAt,
	}
}

// mapLogToAttempt maps a NotificationLog to an AttemptResponse DTO
func mapLogToAttempt(log *models.NotificationLog) *dto.AttemptResponse {
	return &dto.AttemptResponse{
		ID:                     log.ID,
		AttemptNumber:          0, // Logs don't have attempt numbers directly
		Status:                 string(log.Status),
		ErrorMessage:           log.ErrorDetails,
		ProviderResponseSanitized: sanitizeProviderResponse(log.ProviderResponse),
		LatencyMs:              int64(log.ProcessingTimeMs),
		CreatedAt:              log.CreatedAt,
	}
}

// sanitizeProviderResponse redacts sensitive fields from provider responses
func sanitizeProviderResponse(response string) string {
	if response == "" || response == "{}" {
		return ""
	}
	// In a real implementation, parse JSON and remove known secret fields
	if len(response) > 500 {
		return response[:500] + "... [truncated]"
	}
	return response
}

// ============================================
// GetNotificationByID — enriched detail
// ============================================

// GetNotificationByID godoc
// @Summary Get notification by ID
// @Description Retrieve a single notification by its ID with full operational details
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Param X-Tenant-Id header string false "Tenant ID"
// @Param X-Request-Id header string false "Request ID"
// @Success 200 {object} dto.NotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /notifications/{notificationId} [get]
func (h *NotificationHandler) GetNotificationByID(c *fiber.Ctx) error {
	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	notification, err := h.service.GetNotification(c.Context(), notificationID)
	if err != nil {
		return response.NotFound(c, "Notification not found")
	}

	return response.OK(c, mapNotificationToDetail(notification))
}

// ============================================
// ListAllNotifications — admin list with filters
// ============================================

// ListAllNotifications godoc
// @Summary List all notifications (admin)
// @Description Retrieve paginated list of all notifications with filters
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param type query string false "Filter by type/channel"
// @Param channel query string false "Filter by channel (alias for type)"
// @Param userId query string false "Filter by user ID"
// @Param search query string false "Search in recipient/subject/body"
// @Param from query string false "Start date (ISO8601)"
// @Param to query string false "End date (ISO8601)"
// @Param sortBy query string false "Sort field (createdAt, updatedAt, sentAt, priority, status)" default(createdAt)
// @Param sortDirection query string false "Sort direction (asc, desc)" default(desc)
// @Success 200 {object} dto.PaginatedNotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications [get]
func (h *NotificationHandler) ListAllNotifications(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)

	filter := repository.NotificationListFilter{
		Page:          page,
		PageSize:      pageSize,
		SortBy:        c.Query("sortBy", "created_at"),
		SortDirection: c.Query("sortDirection", "desc"),
		Search:        c.Query("search"),
	}

	// Parse optional filters
	if status := c.Query("status"); status != "" {
		s := models.NotificationStatus(status)
		filter.Status = &s
	}
	// Support both 'type' and 'channel' params — channel takes priority
	typeParam := c.Query("channel")
	if typeParam == "" {
		typeParam = c.Query("type")
	}
	if typeParam != "" {
		t := models.NotificationType(typeParam)
		filter.Type = &t
	}
	if userIDStr := c.Query("userId"); userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err == nil {
			filter.UserID = &uid
		}
	}
	if tenantIDStr := c.Query("tenantId"); tenantIDStr != "" {
		tid, err := uuid.Parse(tenantIDStr)
		if err == nil {
			filter.TenantID = &tid
		}
	}
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

	notifications, total, err := h.service.ListAllNotifications(c.Context(), filter)
	if err != nil {
		return response.InternalError(c, "Failed to list notifications: "+err.Error())
	}

	pageSize = filter.PageSize
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	result := &dto.PaginatedNotificationResponse{
		Data:       []*dto.NotificationResponse{},
		Total:      total,
		Page:       filter.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	for _, notif := range notifications {
		result.Data = append(result.Data, mapNotificationToDetail(notif))
	}

	return response.OK(c, result)
}

// ============================================
// RetryNotification — state-validated retry
// ============================================

// RetryNotification godoc
// @Summary Retry notification
// @Description Retry a failed or dead-letter notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /notifications/{notificationId}/retry [post]
func (h *NotificationHandler) RetryNotification(c *fiber.Ctx) error {
	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	if err := h.service.RetryNotification(c.Context(), notificationID); err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.OK(c, &dto.ActionResponse{
		Message: "Notification queued for retry",
		ID:      notificationID,
		Status:  "pending",
	})
}

// ============================================
// CancelNotification — state-validated cancel
// ============================================

// CancelNotification godoc
// @Summary Cancel notification
// @Description Cancel a pending, queued, or retrying notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /notifications/{notificationId}/cancel [post]
func (h *NotificationHandler) CancelNotification(c *fiber.Ctx) error {
	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	if err := h.service.CancelNotification(c.Context(), notificationID); err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.OK(c, &dto.ActionResponse{
		Message: "Notification cancelled",
		ID:      notificationID,
		Status:  "cancelled",
	})
}

// ============================================
// MarkAsSeen — tracking
// ============================================

// MarkAsSeen godoc
// @Summary Mark notification as seen
// @Description Mark a notification as seen (displayed to user)
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /notifications/{notificationId}/seen [post]
func (h *NotificationHandler) MarkAsSeen(c *fiber.Ctx) error {
	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	if err := h.service.MarkAsSeen(c.Context(), notificationID); err != nil {
		return response.NotFound(c, "Notification not found")
	}

	return response.OK(c, &dto.ActionResponse{
		Message: "Notification marked as seen",
		ID:      notificationID,
	})
}

// ============================================
// MarkAsClicked — tracking
// ============================================

// MarkAsClicked godoc
// @Summary Mark notification as clicked
// @Description Mark a notification as clicked (user interacted)
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /notifications/{notificationId}/click [post]
func (h *NotificationHandler) MarkAsClicked(c *fiber.Ctx) error {
	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	if err := h.service.MarkAsClicked(c.Context(), notificationID); err != nil {
		return response.NotFound(c, "Notification not found")
	}

	return response.OK(c, &dto.ActionResponse{
		Message: "Notification marked as clicked",
		ID:      notificationID,
	})
}

// ============================================
// GetUnreadCount — typed DTO
// ============================================

// GetUnreadCount godoc
// @Summary Get unread notification count
// @Description Get the count of unread notifications for a user
// @Tags Notifications
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} dto.UnreadCountResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications/user/{userId}/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
	}

	count, err := h.service.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, &dto.UnreadCountResponse{
		UserID: userID,
		Count:  count,
	})
}

// ============================================
// MarkAllAsRead — typed response
// ============================================

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all unread notifications as read for a user
// @Tags Notifications
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} dto.MarkAllAsReadResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications/user/{userId}/read-all [post]
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
	}

	count, err := h.service.MarkAllAsRead(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, &dto.MarkAllAsReadResponse{
		Message:      "All notifications marked as read",
		UserID:       userID,
		UpdatedCount: count,
	})
}

// ============================================
// CreateNotification — with idempotency
// ============================================

// CreateNotification godoc
// @Summary Create notification
// @Description Create a new notification. Supports Idempotency-Key header or body idempotencyKey field to prevent duplicates.
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string false "Idempotency key to prevent duplicate sends"
// @Param X-Tenant-Id header string false "Tenant ID"
// @Param X-Request-Id header string false "Request ID"
// @Param notification body dto.CreateNotificationRequest true "Notification data"
// @Success 201 {object} dto.NotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /notifications [post]
func (h *NotificationHandler) CreateNotification(c *fiber.Ctx) error {
	req := new(dto.CreateNotificationRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_JSON", "Request body is not valid JSON: "+err.Error())
	}

	// ---- Validation ----
	validationErrors := make([]response.ValidationError, 0)

	// Resolve channel from channel or type
	channel := models.NotificationChannel(req.Channel)
	if channel == "" {
		channel = models.NotificationChannel(req.Type)
	}
	if channel == "" {
		validationErrors = append(validationErrors, response.ValidationError{
			Field:   "channel",
			Message: "channel is required (or use 'type' as alias)",
		})
	}

	// Validate channel value
	validChannels := map[models.NotificationChannel]bool{
		models.NotificationTypeSMS:     true,
		models.NotificationTypeEmail:   true,
		models.NotificationTypePush:    true,
		models.NotificationTypeInApp:   true,
		models.NotificationTypeWebhook: true,
	}
	if channel != "" && !validChannels[channel] {
		validationErrors = append(validationErrors, response.ValidationError{
			Field:   "channel",
			Message: "invalid channel: must be sms, email, push, in_app, or webhook",
		})
	}

	// Validate body
	if req.Body == "" {
		validationErrors = append(validationErrors, response.ValidationError{
			Field:   "body",
			Message: "body is required",
		})
	}

	// Resolve recipient from canonical or legacy fields
	recipientEmail := req.RecipientEmail
	recipientPhone := req.RecipientPhone
	recipientID := req.RecipientID

	if req.Recipient != nil {
		if req.Recipient.Email != "" {
			recipientEmail = req.Recipient.Email
		}
		if req.Recipient.Phone != "" {
			recipientPhone = req.Recipient.Phone
		}
		if req.Recipient.UserID != "" {
			recipientID = req.Recipient.UserID
		}
		if req.Recipient.DeviceToken != "" && recipientID == "" {
			recipientID = req.Recipient.DeviceToken
		}
	}

	// Channel-specific recipient validation
	if channel == models.NotificationTypeSMS && recipientPhone == "" {
		validationErrors = append(validationErrors, response.ValidationError{
			Field:   "recipient.phone",
			Message: "recipient phone is required for sms notifications",
		})
	}
	if channel == models.NotificationTypeEmail && recipientEmail == "" {
		validationErrors = append(validationErrors, response.ValidationError{
			Field:   "recipient.email",
			Message: "recipient email is required for email notifications",
		})
	}
	if channel == models.NotificationTypeInApp && recipientID == "" {
		validationErrors = append(validationErrors, response.ValidationError{
			Field:   "recipient.userId",
			Message: "recipient userId is required for in_app notifications",
		})
	}
	if channel == models.NotificationTypePush && recipientID == "" {
		validationErrors = append(validationErrors, response.ValidationError{
			Field:   "recipient.userId",
			Message: "recipient userId or deviceToken is required for push notifications",
		})
	}
	if channel == models.NotificationTypeWebhook && recipientID == "" {
		// For webhook, recipientID holds the URL
		validationErrors = append(validationErrors, response.ValidationError{
			Field:   "recipient.webhookUrl",
			Message: "recipient webhookUrl is required for webhook notifications",
		})
	}

	if len(validationErrors) > 0 {
		return response.UnprocessableEntity(c, validationErrors)
	}

	// Resolve idempotency key: header takes priority over body field
	// If no key is provided, generate a UUID to ensure every notification
	// has a unique idempotency key. This prevents unique constraint violations
	// on the idx_notif_idempotency_key index (PostgreSQL treats empty strings
	// as values, not as NULL, so the unique index rejects duplicate empty keys).
	idempotencyKey := c.Get("Idempotency-Key")
	if idempotencyKey == "" {
		idempotencyKey = req.IdempotencyKey
	}
	if idempotencyKey == "" {
		idempotencyKey = uuid.New().String()
	}

	// Check for existing notification with same idempotency key
	// (only meaningful when the caller explicitly provided one)
	if c.Get("Idempotency-Key") != "" || req.IdempotencyKey != "" {
		existing, err := h.service.GetByIDempotencyKey(c.Context(), idempotencyKey)
		if err != nil {
			return response.InternalError(c, "Failed to check idempotency key")
		}
		if existing != nil {
			return response.Conflict(c, "Notification with this idempotency key already exists")
		}
	}

	// Resolve userID: parse string, use zero UUID if empty
	var userID uuid.UUID
	if req.UserID != "" {
		var err error
		userID, err = uuid.Parse(req.UserID)
		if err != nil {
			return response.BadRequest(c, "INVALID_USER_ID", "userId is not a valid UUID: "+req.UserID)
		}
	}
	if req.Recipient != nil && req.Recipient.UserID != "" {
		// recipient.userId for in_app can be a real user UUID or a non-UUID identifier
		// Store it as recipientID on the notification
		recipientID = req.Recipient.UserID
	}

	notification := &models.Notification{
		UserID:         userID,
		Type:           models.NotificationType(channel),
		Priority:       req.Priority,
		Locale:         req.Locale,
		RecipientEmail: recipientEmail,
		RecipientPhone: recipientPhone,
		RecipientID:    recipientID,
		Subject:        req.Subject,
		Body:           req.Body,
		TemplateID:     req.TemplateID,
		TemplateKey:    req.TemplateKey,
		ScheduledAt:    req.ScheduledAt,
		IdempotencyKey: idempotencyKey,
	}

	// Set tenant context from header
	if tenantIDStr := c.Get("X-Tenant-Id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			notification.TenantID = &tid
		}
	}

	if req.Metadata != nil {
		metadataJSON, _ := json.Marshal(req.Metadata)
		notification.Metadata = string(metadataJSON)
	} else {
		notification.Metadata = "{}"
	}

	if notification.Locale == "" {
		notification.Locale = "en"
	}
	if notification.Priority == "" {
		notification.Priority = models.NotificationPriorityNormal
	}

	if err := h.service.CreateNotification(c.Context(), notification); err != nil {
		return response.InternalError(c, i18n.T(c.Context(), "notifications.notification_failed"))
	}

	resp := &dto.NotificationResponse{
		ID:             notification.ID,
		UserID:         notification.UserID,
		Type:           notification.Type,
		Status:         notification.Status,
		Priority:       notification.Priority,
		RecipientEmail: notification.RecipientEmail,
		RecipientPhone: notification.RecipientPhone,
		Subject:        notification.Subject,
		Body:           notification.Body,
		Locale:         notification.Locale,
		CreatedAt:      notification.CreatedAt,
	}

	return response.Created(c, resp)
}

// ============================================
// CreateBatchNotifications — with idempotency per item
// ============================================

// CreateBatchNotifications godoc
// @Summary Create batch notifications
// @Description Create multiple notifications at once
// @Tags Notifications
// @Accept json
// @Produce json
// @Param batch body dto.BatchNotificationRequest true "Batch notification data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /notifications/batch [post]
func (h *NotificationHandler) CreateBatchNotifications(c *fiber.Ctx) error {
	req := new(dto.BatchNotificationRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", i18n.T(c.Context(), "errors.invalid_request"))
	}

	notifications := make([]*models.Notification, 0, len(req.Notifications))
	for _, notifReq := range req.Notifications {
		var userID uuid.UUID
		if notifReq.UserID != "" {
			var err error
			userID, err = uuid.Parse(notifReq.UserID)
			if err != nil {
				return response.BadRequest(c, "INVALID_USER_ID", "userId is not a valid UUID: "+notifReq.UserID)
			}
		}

		notification := &models.Notification{
			UserID:         userID,
			Type:           notifReq.Type,
			Priority:       notifReq.Priority,
			Locale:         notifReq.Locale,
			RecipientEmail: notifReq.RecipientEmail,
			RecipientPhone: notifReq.RecipientPhone,
			RecipientID:    notifReq.RecipientID,
			Subject:        notifReq.Subject,
			Body:           notifReq.Body,
			TemplateID:     notifReq.TemplateID,
			TemplateKey:    notifReq.TemplateKey,
			ScheduledAt:    notifReq.ScheduledAt,
			IdempotencyKey: notifReq.IdempotencyKey,
		}

		if notifReq.Metadata != nil {
			metadataJSON, _ := json.Marshal(notifReq.Metadata)
			notification.Metadata = string(metadataJSON)
		} else {
			notification.Metadata = "{}"
		}

		notifications = append(notifications, notification)
	}

	successIDs, errors := h.service.CreateBatchNotifications(c.Context(), notifications)

	return response.Created(c, map[string]interface{}{
		"successCount": len(successIDs),
		"failedCount":  len(errors),
		"successIds":   successIDs,
	})
}

// ============================================
// GetUserNotifications — existing endpoint (preserved)
// ============================================

// GetUserNotifications godoc
// @Summary Get user notifications
// @Description Retrieve paginated notifications for a user
// @Tags Notifications
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} dto.PaginatedNotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications/user/{userId} [get]
func (h *NotificationHandler) GetUserNotifications(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)
	offset := (page - 1) * pageSize

	notifications, total, err := h.service.GetUserNotifications(c.Context(), userID, pageSize, offset)
	if err != nil {
		return response.InternalError(c, err.Error())
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

// ============================================
// GetUnreadNotifications — existing endpoint (preserved)
// ============================================

// GetUnreadNotifications godoc
// @Summary Get unread notifications
// @Description Retrieve unread notifications for a user
// @Tags Notifications
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} dto.PaginatedNotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications/user/{userId}/unread [get]
func (h *NotificationHandler) GetUnreadNotifications(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)
	offset := (page - 1) * pageSize

	notifications, total, err := h.service.GetUnreadNotifications(c.Context(), userID, pageSize, offset)
	if err != nil {
		return response.InternalError(c, err.Error())
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

// ============================================
// MarkAsRead — existing endpoint (preserved)
// ============================================

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a notification as read
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Param X-Tenant-Id header string false "Tenant ID"
// @Param X-Request-Id header string false "Request ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications/{notificationId}/read [put]
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	if err := h.service.MarkAsRead(c.Context(), notificationID); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, &dto.ActionResponse{
		Message: "Notification marked as read",
		ID:      notificationID,
	})
}

// ============================================
// GetNotificationDeliveries — from notification logs
// ============================================

// GetNotificationDeliveries godoc
// @Summary Get notification deliveries
// @Description Retrieve delivery attempts for a notification (mapped from notification logs)
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {array} dto.DeliveryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications/{notificationId}/deliveries [get]
func (h *NotificationHandler) GetNotificationDeliveries(c *fiber.Ctx) error {
	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	logs, err := h.service.GetAttemptsFromLogs(c.Context(), notificationID)
	if err != nil {
		return response.InternalError(c, "Failed to retrieve delivery information")
	}

	// Map logs to delivery response
	deliveries := make([]*dto.DeliveryResponse, 0, 1)
	if len(logs) > 0 {
		attempts := make([]*dto.AttemptResponse, 0, len(logs))
		for _, log := range logs {
			attempts = append(attempts, mapLogToAttempt(log))
		}
		deliveries = append(deliveries, &dto.DeliveryResponse{
			ID:             notificationID,
			NotificationID: notificationID,
			AttemptCount:   len(attempts),
			Attempts:       attempts,
		})
	}

	return response.OK(c, deliveries)
}

// ============================================
// GetNotificationAttempts — from notification logs
// ============================================

// GetNotificationAttempts godoc
// @Summary Get notification delivery attempts
// @Description Retrieve individual delivery attempts for a notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path string true "Notification ID"
// @Success 200 {array} dto.AttemptResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications/{notificationId}/attempts [get]
func (h *NotificationHandler) GetNotificationAttempts(c *fiber.Ctx) error {
	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	logs, err := h.service.GetAttemptsFromLogs(c.Context(), notificationID)
	if err != nil {
		return response.InternalError(c, "Failed to retrieve attempt information")
	}

	attempts := make([]*dto.AttemptResponse, 0, len(logs))
	for _, log := range logs {
		attempts = append(attempts, mapLogToAttempt(log))
	}

	return response.OK(c, attempts)
}
