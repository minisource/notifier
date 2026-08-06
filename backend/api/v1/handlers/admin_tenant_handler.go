package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

// AdminTenantHandler provides admin/operator tenant management endpoints.
type AdminTenantHandler struct {
	tenantRepo repository.TenantRepository
}

func NewAdminTenantHandler(tenantRepo repository.TenantRepository) *AdminTenantHandler {
	return &AdminTenantHandler{tenantRepo: tenantRepo}
}

type tenantCreateRequest struct {
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	DisplayName     string   `json:"displayName,omitempty"`
	Description     string   `json:"description,omitempty"`
	EnabledChannels []string `json:"enabledChannels,omitempty"`
}

type tenantUpdateRequest struct {
	Name            *string  `json:"name,omitempty"`
	Slug            *string  `json:"slug,omitempty"`
	DisplayName     *string  `json:"displayName,omitempty"`
	Description     *string  `json:"description,omitempty"`
	EnabledChannels []string `json:"enabledChannels,omitempty"`
}

// ListTenants godoc
// @Summary Admin: List tenants
// @Description Retrieve paginated list of all tenants (admin only)
// @Tags Admin Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /admin/tenants [get]
func (h *AdminTenantHandler) ListTenants(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))

	tenants, total, err := h.tenantRepo.List(c.Context(), page, pageSize)
	if err != nil {
		return response.New().Status(fiber.StatusInternalServerError).Error(
			"TENANT_LIST_FAILED", "Failed to list tenants").Send(c)
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return response.New().Data(fiber.Map{
		"data": tenants,
		"meta": fiber.Map{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	}).Send(c)
}

// GetTenant godoc
// @Summary Admin: Get tenant by ID
// @Description Retrieve a single tenant (admin only)
// @Tags Admin Tenants
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tenant ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/tenants/{id} [get]
func (h *AdminTenantHandler) GetTenant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.New().Status(fiber.StatusBadRequest).Error(
			"INVALID_TENANT_ID", "Invalid tenant ID").Send(c)
	}

	tenant, err := h.tenantRepo.GetByID(c.Context(), id)
	if err != nil {
		return response.New().Status(fiber.StatusInternalServerError).Error(
			"TENANT_GET_FAILED", "Failed to get tenant").Send(c)
	}
	if tenant == nil {
		return response.New().Status(fiber.StatusNotFound).Error(
			"TENANT_NOT_FOUND", "Tenant not found").Send(c)
	}

	return response.New().Data(tenant).Send(c)
}

// CreateTenant godoc
// @Summary Admin: Create tenant
// @Description Create a new tenant (admin only)
// @Tags Admin Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body tenantCreateRequest true "Tenant payload"
// @Success 201 {object} map[string]interface{}
// @Router /admin/tenants [post]
func (h *AdminTenantHandler) CreateTenant(c *fiber.Ctx) error {
	var req tenantCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.New().Status(fiber.StatusBadRequest).Error(
			"INVALID_REQUEST", "Invalid request body").Send(c)
	}
	if req.Name == "" || req.Slug == "" {
		return response.New().Status(fiber.StatusBadRequest).Error(
			"VALIDATION_ERROR", "name and slug are required").Send(c)
	}

	// Slug must be unique
	existing, err := h.tenantRepo.GetBySlug(c.Context(), req.Slug)
	if err != nil {
		return response.New().Status(fiber.StatusInternalServerError).Error(
			"TENANT_CREATE_FAILED", "Failed to create tenant").Send(c)
	}
	if existing != nil {
		return response.New().Status(fiber.StatusConflict).Error(
			"TENANT_SLUG_EXISTS", "A tenant with this slug already exists").Send(c)
	}

	tenant := &models.Tenant{
		Name:            req.Name,
		Slug:            req.Slug,
		DisplayName:     req.DisplayName,
		Description:     req.Description,
		IsActive:        true,
		EnabledChannels: req.EnabledChannels,
	}

	if err := h.tenantRepo.Create(c.Context(), tenant); err != nil {
		return response.New().Status(fiber.StatusInternalServerError).Error(
			"TENANT_CREATE_FAILED", "Failed to create tenant").Send(c)
	}

	return response.New().Status(fiber.StatusCreated).Data(tenant).Send(c)
}

// UpdateTenant godoc
// @Summary Admin: Update tenant
// @Description Update tenant fields (admin only)
// @Tags Admin Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tenant ID"
// @Param body body tenantUpdateRequest true "Tenant updates"
// @Success 200 {object} map[string]interface{}
// @Router /admin/tenants/{id} [put]
func (h *AdminTenantHandler) UpdateTenant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.New().Status(fiber.StatusBadRequest).Error(
			"INVALID_TENANT_ID", "Invalid tenant ID").Send(c)
	}

	tenant, err := h.tenantRepo.GetByID(c.Context(), id)
	if err != nil {
		return response.New().Status(fiber.StatusInternalServerError).Error(
			"TENANT_GET_FAILED", "Failed to get tenant").Send(c)
	}
	if tenant == nil {
		return response.New().Status(fiber.StatusNotFound).Error(
			"TENANT_NOT_FOUND", "Tenant not found").Send(c)
	}

	var req tenantUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.New().Status(fiber.StatusBadRequest).Error(
			"INVALID_REQUEST", "Invalid request body").Send(c)
	}

	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.Slug != nil {
		existing, err := h.tenantRepo.GetBySlug(c.Context(), *req.Slug)
		if err == nil && existing != nil && existing.ID != tenant.ID {
			return response.New().Status(fiber.StatusConflict).Error(
				"TENANT_SLUG_EXISTS", "A tenant with this slug already exists").Send(c)
		}
		tenant.Slug = *req.Slug
	}
	if req.DisplayName != nil {
		tenant.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		tenant.Description = *req.Description
	}
	if req.EnabledChannels != nil {
		tenant.EnabledChannels = req.EnabledChannels
	}

	if err := h.tenantRepo.Update(c.Context(), tenant); err != nil {
		return response.New().Status(fiber.StatusInternalServerError).Error(
			"TENANT_UPDATE_FAILED", "Failed to update tenant").Send(c)
	}

	return response.New().Data(tenant).Send(c)
}

// DeleteTenant godoc
// @Summary Admin: Delete tenant
// @Description Delete a tenant (admin only)
// @Tags Admin Tenants
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tenant ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/tenants/{id} [delete]
func (h *AdminTenantHandler) DeleteTenant(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.New().Status(fiber.StatusBadRequest).Error(
			"INVALID_TENANT_ID", "Invalid tenant ID").Send(c)
	}

	tenant, err := h.tenantRepo.GetByID(c.Context(), id)
	if err != nil || tenant == nil {
		return response.New().Status(fiber.StatusNotFound).Error(
			"TENANT_NOT_FOUND", "Tenant not found").Send(c)
	}

	if tenant.IsDefault {
		return response.New().Status(fiber.StatusBadRequest).Error(
			"CANNOT_DELETE_DEFAULT_TENANT", "The default tenant cannot be deleted").Send(c)
	}

	if err := h.tenantRepo.Delete(c.Context(), id); err != nil {
		return response.New().Status(fiber.StatusInternalServerError).Error(
			"TENANT_DELETE_FAILED", "Failed to delete tenant").Send(c)
	}

	return response.New().Data(fiber.Map{"success": true, "id": id.String()}).Send(c)
}

// ToggleTenantStatus godoc
// @Summary Admin: Toggle tenant status
// @Description Activate or deactivate a tenant (admin only)
// @Tags Admin Tenants
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tenant ID"
// @Param status path string true "Status: active|inactive"
// @Success 200 {object} map[string]interface{}
// @Router /admin/tenants/{id}/status/{status} [patch]
func (h *AdminTenantHandler) ToggleTenantStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.New().Status(fiber.StatusBadRequest).Error(
			"INVALID_TENANT_ID", "Invalid tenant ID").Send(c)
	}

	status := c.Params("status")
	if status != "active" && status != "inactive" {
		return response.New().Status(fiber.StatusBadRequest).Error(
			"INVALID_STATUS", "Status must be 'active' or 'inactive'").Send(c)
	}

	tenant, err := h.tenantRepo.GetByID(c.Context(), id)
	if err != nil || tenant == nil {
		return response.New().Status(fiber.StatusNotFound).Error(
			"TENANT_NOT_FOUND", "Tenant not found").Send(c)
	}

	isActive := status == "active"
	if err := h.tenantRepo.SetStatus(c.Context(), id, isActive); err != nil {
		return response.New().Status(fiber.StatusInternalServerError).Error(
			"TENANT_UPDATE_FAILED", "Failed to update tenant").Send(c)
	}

	tenant.IsActive = isActive
	return response.New().Data(tenant).Send(c)
}
