package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

type TemplateHandler struct {
	repo repository.NotificationTemplateRepository
}

func NewTemplateHandler(repo repository.NotificationTemplateRepository) *TemplateHandler {
	return &TemplateHandler{repo: repo}
}

// CreateTemplate godoc
// @Summary Create notification template
// @Description Create a new notification template
// @Tags Templates
// @Accept json
// @Produce json
// @Param template body dto.CreateTemplateRequest true "Template data"
// @Success 201 {object} dto.TemplateResponse
// @Failure 400 {object} map[string]interface{}
// @Router /v1/templates [post]
func (h *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	req := new(dto.CreateTemplateRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", i18n.T(c.Context(), "errors.invalid_request"))
	}

	template := &models.NotificationTemplate{
		Name:             req.Name,
		Type:             req.Type,
		Subject:          req.Subject,
		Body:             req.Body,
		Description:      req.Description,
		Provider:         req.Provider,
		ProviderTemplate: req.ProviderTemplate,
		IsActive:         true,
	}

	if req.Variables != nil {
		variablesJSON, _ := json.Marshal(req.Variables)
		template.Variables = string(variablesJSON)
	}

	if err := h.repo.Create(c.Context(), template); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, template)
}

// GetAllTemplates godoc
// @Summary Get all templates
// @Description Retrieve all notification templates
// @Tags Templates
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(50)
// @Success 200 {object} map[string]interface{}
// @Router /v1/templates [get]
func (h *TemplateHandler) GetAllTemplates(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 50)
	offset := (page - 1) * pageSize

	templates, total, err := h.repo.GetAll(c.Context(), pageSize, offset)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, map[string]interface{}{
		"data":  templates,
		"total": total,
		"page":  page,
	})
}

// GetTemplate godoc
// @Summary Get template by ID
// @Description Retrieve a specific notification template
// @Tags Templates
// @Accept json
// @Produce json
// @Param templateId path string true "Template ID"
// @Success 200 {object} dto.TemplateResponse
// @Failure 404 {object} map[string]interface{}
// @Router /v1/templates/{templateId} [get]
func (h *TemplateHandler) GetTemplate(c *fiber.Ctx) error {
	templateIDStr := c.Params("templateId")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_TEMPLATE_ID", "Invalid template ID")
	}

	template, err := h.repo.GetByID(c.Context(), templateID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, template)
}

// UpdateTemplate godoc
// @Summary Update template
// @Description Update an existing notification template
// @Tags Templates
// @Accept json
// @Produce json
// @Param templateId path string true "Template ID"
// @Param template body dto.CreateTemplateRequest true "Template data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/templates/{templateId} [put]
func (h *TemplateHandler) UpdateTemplate(c *fiber.Ctx) error {
	templateIDStr := c.Params("templateId")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_TEMPLATE_ID", "Invalid template ID")
	}

	req := new(dto.CreateTemplateRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", i18n.T(c.Context(), "errors.invalid_request"))
	}

	template, err := h.repo.GetByID(c.Context(), templateID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	template.Name = req.Name
	template.Type = req.Type
	template.Subject = req.Subject
	template.Body = req.Body
	template.Description = req.Description
	template.Provider = req.Provider
	template.ProviderTemplate = req.ProviderTemplate

	if req.Variables != nil {
		variablesJSON, _ := json.Marshal(req.Variables)
		template.Variables = string(variablesJSON)
	}

	if err := h.repo.Update(c.Context(), template); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, template)
}

// DeleteTemplate godoc
// @Summary Delete template
// @Description Delete a notification template
// @Tags Templates
// @Accept json
// @Produce json
// @Param templateId path string true "Template ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/templates/{templateId} [delete]
func (h *TemplateHandler) DeleteTemplate(c *fiber.Ctx) error {
	templateIDStr := c.Params("templateId")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_TEMPLATE_ID", "Invalid template ID")
	}

	if err := h.repo.Delete(c.Context(), templateID); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{"message": "Template deleted successfully"})
}
