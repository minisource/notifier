package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/service"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(service *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// CreateNotification godoc
// @Summary Create notification
// @Description Create a new notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notification body dto.CreateNotificationRequest true "Notification data"
// @Success 201 {object} dto.NotificationResponse
// @Failure 400 {object} map[string]interface{}
// @Router /v1/notifications [post]
func (h *NotificationHandler) CreateNotification(c *fiber.Ctx) error {
	req := new(dto.CreateNotificationRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", i18n.T(c.Context(), "errors.invalid_request"))
	}

	notification := &models.Notification{
		UserID:         req.UserID,
		Type:           req.Type,
		Priority:       req.Priority,
		RecipientEmail: req.RecipientEmail,
		RecipientPhone: req.RecipientPhone,
		RecipientID:    req.RecipientID,
		Subject:        req.Subject,
		Body:           req.Body,
		TemplateID:     req.TemplateID,
		ScheduledAt:    req.ScheduledAt,
	}

	if req.Metadata != nil {
		metadataJSON, _ := json.Marshal(req.Metadata)
		notification.Metadata = string(metadataJSON)
	}

	if err := h.service.CreateNotification(c.Context(), notification); err != nil {
		return response.InternalError(c, i18n.T(c.Context(), "notifications.notification_failed"))
	}

	return response.Created(c, map[string]interface{}{
		"id":      notification.ID,
		"message": i18n.T(c.Context(), "notifications.notification_created"),
	})
}

// CreateBatchNotifications godoc
// @Summary Create batch notifications
// @Description Create multiple notifications at once
// @Tags Notifications
// @Accept json
// @Produce json
// @Param batch body dto.BatchNotificationRequest true "Batch notification data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/notifications/batch [post]
func (h *NotificationHandler) CreateBatchNotifications(c *fiber.Ctx) error {
	req := new(dto.BatchNotificationRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", i18n.T(c.Context(), "errors.invalid_request"))
	}

	notifications := make([]*models.Notification, 0, len(req.Notifications))
	for _, notifReq := range req.Notifications {
		notification := &models.Notification{
			UserID:         notifReq.UserID,
			Type:           notifReq.Type,
			Priority:       notifReq.Priority,
			RecipientEmail: notifReq.RecipientEmail,
			RecipientPhone: notifReq.RecipientPhone,
			RecipientID:    notifReq.RecipientID,
			Subject:        notifReq.Subject,
			Body:           notifReq.Body,
			TemplateID:     notifReq.TemplateID,
			ScheduledAt:    notifReq.ScheduledAt,
		}

		if notifReq.Metadata != nil {
			metadataJSON, _ := json.Marshal(notifReq.Metadata)
			notification.Metadata = string(metadataJSON)
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
// @Failure 400 {object} map[string]interface{}
// @Router /v1/notifications/user/{userId} [get]
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
		result.Data = append(result.Data, &dto.NotificationResponse{
			ID:             notif.ID,
			UserID:         notif.UserID,
			Type:           notif.Type,
			Status:         notif.Status,
			Priority:       notif.Priority,
			RecipientEmail: notif.RecipientEmail,
			RecipientPhone: notif.RecipientPhone,
			Subject:        notif.Subject,
			Body:           notif.Body,
			ReadAt:         notif.ReadAt,
			SentAt:         notif.SentAt,
			CreatedAt:      notif.CreatedAt,
		})
	}

	return response.OK(c, result)
}

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
// @Failure 400 {object} map[string]interface{}
// @Router /v1/notifications/user/{userId}/unread [get]
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
		result.Data = append(result.Data, &dto.NotificationResponse{
			ID:             notif.ID,
			UserID:         notif.UserID,
			Type:           notif.Type,
			Status:         notif.Status,
			Priority:       notif.Priority,
			RecipientEmail: notif.RecipientEmail,
			RecipientPhone: notif.RecipientPhone,
			Subject:        notif.Subject,
			Body:           notif.Body,
			CreatedAt:      notif.CreatedAt,
		})
	}

	return response.OK(c, result)
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a notification as read
// @Tags Notifications
// @Accept json
// @Produce json
// @Param notificationId path string true "Notification ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/notifications/{notificationId}/read [put]
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	notificationIDStr := c.Params("notificationId")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_NOTIFICATION_ID", "Invalid notification ID")
	}

	if err := h.service.MarkAsRead(c.Context(), notificationID); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{"message": "Notification marked as read"})
}
