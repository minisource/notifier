package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	helper "github.com/minisource/common_go/http/helpers"
	"github.com/minisource/notifire/api/v1/dto"
	"github.com/minisource/notifire/config"
	"github.com/minisource/notifire/internal/notification"
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
func (h *SMSHandler) SendSMS(c *gin.Context) {
	req := new(dto.SMSRequest)
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			helper.GenerateBaseResponseWithValidationError(nil, false, helper.ValidationError, err))
		return
	}

	err = h.service.SendNotification(*req)
	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusOK, helper.GenerateBaseResponse(nil, true, 0))
}