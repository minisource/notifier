package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/common_go/http/helper"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/notification"
)

type SMSHandler struct {
	service *notification.SMSService
}

func NewSMSHandler(cfg *config.Config) *SMSHandler {
	service := notification.NewSMSService(cfg)
	return &SMSHandler{service: service}
}

// SendSMS godoc
// @Summary Send SMS
// @Description Send SMS
// @Tags SMS
// @Accept  json
// @Produce  json
// @Success 200 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Router /v1/Sms/ [post]
func (h *SMSHandler) SendSMS(c *fiber.Ctx) error {
	// Parse the request body
	req := new(dto.SMSRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.GenerateBaseResponseWithValidationError(nil, false, helper.ValidationError, err),
		)
	}

	// Call the service to send the SMS
	err := h.service.SendNotification(*req)
	if err != nil {
		return c.Status(helper.TranslateErrorToStatusCode(err)).JSON(
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err),
		)
	}

	// Return a success response
	return c.Status(fiber.StatusOK).JSON(
		helper.GenerateBaseResponse(nil, true, 0),
	)
}
