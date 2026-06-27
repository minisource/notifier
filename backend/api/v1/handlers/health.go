package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/response"
)

type HealthHandler struct {
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck godoc
// @Summary Health Check
// @Description Health Check
// @Tags health
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{} "Success"
// @Failure 400 {object} map[string]interface{} "Failed"
// @Router /health/ [get]
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return response.OK(c, "Working!")
}
