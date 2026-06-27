package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/internal/repository"
)

// InAppHandler handles in-app notification operations (inbox, tracking)
type InAppHandler struct {
	notifRepo repository.NotificationRepository
}

// NewInAppHandler creates a new in-app handler
func NewInAppHandler(notifRepo repository.NotificationRepository) *InAppHandler {
	return &InAppHandler{notifRepo: notifRepo}
}

// GetInAppNotifications returns in-app notifications for a user with pagination
func (h *InAppHandler) GetInAppNotifications(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_USER_ID", "Invalid user ID")
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)
	offset := (page - 1) * pageSize

	notifications, total, err := h.notifRepo.GetInAppByUserID(c.Context(), userID, pageSize, offset)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return response.OK(c, fiber.Map{
		"data":       notifications,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	})
}

// MarkAsSeen marks a notification as seen (displayed in user's inbox)
func (h *InAppHandler) MarkAsSeen(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid notification ID")
	}

	if err := h.notifRepo.MarkAsSeen(c.Context(), id); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{"message": "Notification marked as seen"})
}

// MarkAsRead marks a notification as read (user opened/viewed content)
func (h *InAppHandler) MarkAsRead(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid notification ID")
	}

	if err := h.notifRepo.MarkAsRead(c.Context(), id); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{"message": "Notification marked as read"})
}

// MarkAsClicked marks a notification as clicked (user tapped/interacted)
func (h *InAppHandler) MarkAsClicked(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid notification ID")
	}

	if err := h.notifRepo.MarkAsClicked(c.Context(), id); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{"message": "Notification marked as clicked"})
}
