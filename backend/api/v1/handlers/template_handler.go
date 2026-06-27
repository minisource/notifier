package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/service"
)

type TemplateHandler struct {
	templateService *service.TemplateService
}

func NewTemplateHandler(templateService *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
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
// @Router /templates [post]
func (h *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	req := new(dto.CreateTemplateRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", i18n.T(c.Context(), "errors.invalid_request"))
	}

	template := &models.NotificationTemplate{
		Key:              req.Key,
		Name:             req.Name,
		Type:             req.Type,
		Locale:           req.Locale,
		Subject:          req.Subject,
		Body:             req.Body,
		Description:      req.Description,
		Provider:         req.Provider,
		ProviderTemplate: req.ProviderTemplate,
		IsActive:         true,
	}

	if req.Variables != nil {
		if err := template.SetVariables(req.Variables); err != nil {
			return response.BadRequest(c, "INVALID_VARIABLES", "Invalid variables format")
		}
	} else {
		template.Variables = "[]"
	}

	result, err := h.templateService.CreateTemplate(c.Context(), template)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, result)
}

// GetAllTemplates godoc
// @Summary Get all templates
// @Description Retrieve all notification templates with pagination and optional filters
// @Tags Templates
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(50)
// @Param channel query string false "Filter by channel/type"
// @Param locale query string false "Filter by locale"
// @Param isActive query bool false "Filter by active status"
// @Success 200 {object} dto.PaginatedResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /templates [get]
func (h *TemplateHandler) GetAllTemplates(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 50)

	templates, total, err := h.templateService.ListTemplates(c.Context(), page, pageSize)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return response.OK(c, &dto.PaginatedResponse{
		Items:      templates,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
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
// @Router /templates/{templateId} [get]
func (h *TemplateHandler) GetTemplate(c *fiber.Ctx) error {
	templateIDStr := c.Params("templateId")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_TEMPLATE_ID", "Invalid template ID")
	}

	template, err := h.templateService.GetTemplate(c.Context(), templateID)
	if err != nil {
		return response.NotFound(c, "Template not found")
	}

	return response.OK(c, template)
}

// GetTemplateByKey godoc
// @Summary Get template by key
// @Description Retrieve a template by its programmatic key (e.g., "auth.otp.sms")
// @Tags Templates
// @Accept json
// @Produce json
// @Param key path string true "Template key"
// @Success 200 {object} dto.TemplateResponse
// @Failure 404 {object} map[string]interface{}
// @Router /templates/key/{key} [get]
func (h *TemplateHandler) GetTemplateByKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return response.BadRequest(c, "INVALID_KEY", "Template key is required")
	}

	template, err := h.templateService.GetTemplateByKey(c.Context(), key)
	if err != nil {
		return response.NotFound(c, "Template not found for key: "+key)
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
// @Router /templates/{templateId} [put]
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

	updates := &models.NotificationTemplate{
		Key:              req.Key,
		Name:             req.Name,
		Type:             req.Type,
		Locale:           req.Locale,
		Subject:          req.Subject,
		Body:             req.Body,
		Description:      req.Description,
		Provider:         req.Provider,
		ProviderTemplate: req.ProviderTemplate,
		IsActive:         req.IsActive,
	}

	if req.Variables != nil {
		if err := updates.SetVariables(req.Variables); err != nil {
			return response.BadRequest(c, "INVALID_VARIABLES", "Invalid variables format")
		}
	}

	result, err := h.templateService.UpdateTemplate(c.Context(), templateID, updates)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, result)
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
// @Router /templates/{templateId} [delete]
func (h *TemplateHandler) DeleteTemplate(c *fiber.Ctx) error {
	templateIDStr := c.Params("templateId")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_TEMPLATE_ID", "Invalid template ID")
	}

	if err := h.templateService.DeleteTemplate(c.Context(), templateID); err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{"message": "Template deleted successfully"})
}

// RenderPreview godoc
// @Summary Render template preview
// @Description Render a template with sample variables to preview the output
// @Tags Templates
// @Accept json
// @Produce json
// @Param templateId path string true "Template ID"
// @Param preview body dto.RenderPreviewRequest true "Preview data"
// @Success 200 {object} dto.RenderPreviewResponse
// @Failure 400 {object} map[string]interface{}
// @Router /templates/{templateId}/render-preview [post]
func (h *TemplateHandler) RenderPreview(c *fiber.Ctx) error {
	templateIDStr := c.Params("templateId")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_TEMPLATE_ID", "Invalid template ID")
	}

	req := new(dto.RenderPreviewRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", i18n.T(c.Context(), "errors.invalid_request"))
	}

	if req.Variables == nil {
		req.Variables = make(map[string]string)
	}

	result, err := h.templateService.RenderPreview(c.Context(), templateID, req.Variables, req.Locale)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, &dto.RenderPreviewResponse{
		Subject:         result.Subject,
		Body:            result.Body,
		UsedVariables:   result.UsedVariables,
		MissingVariables: result.MissingVariables,
	})
}

// RenderPreviewByKey godoc
// @Summary Render template preview by key
// @Description Render a template with sample variables using its programmatic key
// @Tags Templates
// @Accept json
// @Produce json
// @Param preview body dto.RenderPreviewRequest true "Preview data (supports templateKey field)"
// @Success 200 {object} dto.RenderPreviewResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /templates/render-preview [post]
func (h *TemplateHandler) RenderPreviewByKey(c *fiber.Ctx) error {
	req := new(dto.RenderPreviewRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	if req.TemplateKey == "" {
		return response.BadRequest(c, "MISSING_TEMPLATE_KEY", "Template key is required")
	}

	if req.Variables == nil {
		req.Variables = make(map[string]string)
	}

	// Look up template by key
	template, err := h.templateService.GetTemplateByKey(c.Context(), req.TemplateKey)
	if err != nil {
		return response.NotFound(c, "Template not found for key: "+req.TemplateKey)
	}

	// Use locale from request or template default
	locale := req.Locale
	if locale == "" {
		locale = template.Locale
	}

	result, err := service.RenderTemplate(&service.RenderRequest{
		Template:  template,
		Variables: req.Variables,
		Locale:    locale,
	})
	if err != nil {
		return response.InternalError(c, "Failed to render template")
	}

	return response.OK(c, &dto.RenderPreviewResponse{
		Subject:         result.Subject,
		Body:            result.Body,
		UsedVariables:   result.UsedVariables,
		MissingVariables: result.MissingVariables,
	})
}

// UpdateTemplateStatus godoc
// @Summary Update template status
// @Description Activate or deactivate a template
// @Tags Templates
// @Accept json
// @Produce json
// @Param templateId path string true "Template ID"
// @Param status body object true "Status update"
// @Success 200 {object} dto.TemplateResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /templates/{templateId}/status [patch]
func (h *TemplateHandler) UpdateTemplateStatus(c *fiber.Ctx) error {
	templateIDStr := c.Params("templateId")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_TEMPLATE_ID", "Invalid template ID")
	}

	var body struct {
		IsActive bool `json:"isActive"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	// Get existing template
	template, err := h.templateService.GetTemplate(c.Context(), templateID)
	if err != nil {
		return response.NotFound(c, "Template not found")
	}

	// Update isActive
	template.IsActive = body.IsActive

	// Save via service's internal update
	updated, err := h.templateService.UpdateTemplate(c.Context(), templateID, template)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, updated)
}
