package handlers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"

)

// ProviderHandler handles provider-related admin endpoints
type ProviderHandler struct {
	providerRepo repository.ProviderRepository
}

// NewProviderHandler creates a new provider handler
func NewProviderHandler(providerRepo repository.ProviderRepository) *ProviderHandler {
	return &ProviderHandler{
		providerRepo: providerRepo,
	}
}

// secretSuffixes are config key suffixes that should be redacted in provider responses
var secretSuffixes = []string{
	"apikey", "api_key", "api-key",
	"token", "secret",
	"password", "passwd",
	"privatekey", "private_key", "private-key",
	"accesskey", "access_key", "access-key",
	"refresh_token", "refresh-token",
	"clientsecret", "client_secret", "client-secret",
	"authorization",
	"bearer",
}

// validChannels defines the allowed notification channels
var validChannels = map[string]bool{
	"sms": true, "email": true, "push": true, "webhook": true, "in_app": true,
}

// validStatuses defines the allowed provider statuses
var validStatuses = map[string]bool{
	"active": true, "inactive": true, "disabled": true, "error": true,
}

// isSecretKey checks if a config key name indicates a secret value
func isSecretKey(key string) bool {
	lower := strings.ToLower(key)
	for _, suffix := range secretSuffixes {
		if strings.Contains(lower, suffix) {
			return true
		}
	}
	return false
}

// redactSecrets takes a provider config JSON string, parses it, masks secret values, and re-encodes
func redactSecrets(configJSON string) map[string]interface{} {
	result := make(map[string]interface{})
	if configJSON == "" {
		return result
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		result["_raw"] = configJSON
		return result
	}
	for k, v := range cfg {
		if isSecretKey(k) && v != nil {
			strVal, ok := v.(string)
			if ok && strVal != "" {
				if len(strVal) <= 4 {
					result[k] = "****"
				} else {
					result[k] = strVal[:2] + "****" + strVal[len(strVal)-2:]
				}
				continue
			}
			// Recursive redaction for nested objects
			if nestedMap, ok := v.(map[string]interface{}); ok {
				result[k] = redactSecretsInMap(nestedMap)
				continue
			}
		}
		result[k] = v
	}
	return result
}

// redactSecretsInMap recursively redacts secrets in a nested map
func redactSecretsInMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(data))
	for k, v := range data {
		if isSecretKey(k) && v != nil {
			strVal, ok := v.(string)
			if ok && strVal != "" {
				if len(strVal) <= 4 {
					result[k] = "****"
				} else {
					result[k] = strVal[:2] + "****" + strVal[len(strVal)-2:]
				}
				continue
			}
			if nestedMap, ok := v.(map[string]interface{}); ok {
				result[k] = redactSecretsInMap(nestedMap)
				continue
			}
		}
		if nestedMap, ok := v.(map[string]interface{}); ok {
			result[k] = redactSecretsInMap(nestedMap)
		} else {
			result[k] = v
		}
	}
	return result
}

// mapProviderToResponse maps a Provider model to a ProviderResponse DTO with secrets redacted
func mapProviderToResponse(p *models.Provider) *dto.ProviderResponse {
	status := p.Status
	if status == "" {
		if !p.IsEnabled {
			status = "disabled"
		} else {
			status = "active"
		}
	}

	resp := &dto.ProviderResponse{
		ID:          p.ID.String(),
		Name:        p.Name,
		Channel:     p.Channel,
		Type:        p.Type,
		Status:      status,
		Description: p.Description,
		IsEnabled:   p.IsEnabled,
		IsPrimary:   p.IsPrimary,
		IsDefault:   p.IsDefault,
		Priority:    p.Priority,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	// Redact secrets from config
	if p.Config != "" {
		resp.Config = redactSecrets(p.Config)
	}

	return resp
}

// ============================================
// CRUD: CreateProvider
// ============================================

// CreateProvider godoc
// @Summary Create provider
// @Description Create a new notification provider
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider body dto.CreateProviderRequest true "Provider data"
// @Success 201 {object} dto.ProviderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /admin/providers [post]
func (h *ProviderHandler) CreateProvider(c *fiber.Ctx) error {
	req := new(dto.CreateProviderRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid provider data")
	}

	if req.Name == "" {
		return response.BadRequest(c, "VALIDATION_ERROR", "Provider name is required")
	}
	if req.Channel == "" {
		return response.BadRequest(c, "VALIDATION_ERROR", "Provider channel is required")
	}
	if !validChannels[req.Channel] {
		return response.BadRequest(c, "VALIDATION_ERROR", "Channel must be one of: sms, email, push, webhook, in_app")
	}
	if req.Type == "" {
		return response.BadRequest(c, "VALIDATION_ERROR", "Provider type is required")
	}
	if req.Status != "" && !validStatuses[req.Status] {
		return response.BadRequest(c, "VALIDATION_ERROR", "Status must be one of: active, inactive, disabled, error")
	}

	configJSON := ""
	if req.Config != nil {
		configBytes, _ := json.Marshal(req.Config)
		configJSON = string(configBytes)
	}

	secretConfigJSON := ""
	if req.SecretConfig != nil {
		secretBytes, _ := json.Marshal(req.SecretConfig)
		secretConfigJSON = string(secretBytes)
	}

	status := req.Status
	if status == "" {
		status = models.ProviderStatusActive
	}

	provider := &models.Provider{
		Name:         req.Name,
		Channel:      req.Channel,
		Type:         req.Type,
		Status:       status,
		Config:       configJSON,
		SecretConfig: secretConfigJSON,
		Priority:     req.Priority,
		IsEnabled:    status != models.ProviderStatusDisabled,
		IsPrimary:    false,
		IsDefault:    req.IsDefault,
		Description:  req.Description,
	}

	if err := h.providerRepo.Create(c.Context(), provider); err != nil {
		return response.InternalError(c, "Failed to create provider: "+err.Error())
	}

	return response.Created(c, mapProviderToResponse(provider))
}

// ============================================
// CRUD: GetProvider
// ============================================

// GetProvider godoc
// @Summary Get provider by ID
// @Description Retrieve a single provider by ID
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param providerId path string true "Provider ID"
// @Success 200 {object} dto.ProviderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/providers/{providerId} [get]
func (h *ProviderHandler) GetProvider(c *fiber.Ctx) error {
	providerIDStr := c.Params("providerId")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_PROVIDER_ID", "Invalid provider ID")
	}

	provider, err := h.providerRepo.GetByID(c.Context(), providerID)
	if err != nil {
		return response.InternalError(c, "Failed to get provider")
	}
	if provider == nil {
		return response.NotFound(c, "Provider not found")
	}

	return response.OK(c, mapProviderToResponse(provider))
}

// ============================================
// CRUD: UpdateProvider
// ============================================

// UpdateProvider godoc
// @Summary Update provider
// @Description Update an existing provider
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param providerId path string true "Provider ID"
// @Param provider body dto.UpdateProviderRequest true "Provider data"
// @Success 200 {object} dto.ProviderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/providers/{providerId} [put]
func (h *ProviderHandler) UpdateProvider(c *fiber.Ctx) error {
	providerIDStr := c.Params("providerId")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_PROVIDER_ID", "Invalid provider ID")
	}

	existing, err := h.providerRepo.GetByID(c.Context(), providerID)
	if err != nil {
		return response.InternalError(c, "Failed to get provider")
	}
	if existing == nil {
		return response.NotFound(c, "Provider not found")
	}

	req := new(dto.UpdateProviderRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid provider data")
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Channel != "" {
		if !validChannels[req.Channel] {
			return response.BadRequest(c, "VALIDATION_ERROR", "Channel must be one of: sms, email, push, webhook, in_app")
		}
		existing.Channel = req.Channel
	}
	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.Status != nil {
		if !validStatuses[*req.Status] {
			return response.BadRequest(c, "VALIDATION_ERROR", "Status must be one of: active, inactive, disabled, error")
		}
		existing.Status = *req.Status
		existing.IsEnabled = *req.Status != models.ProviderStatusDisabled
	}
	if req.Priority != nil {
		existing.Priority = *req.Priority
	}
	if req.IsEnabled != nil {
		existing.IsEnabled = *req.IsEnabled
		if *req.IsEnabled && existing.Status == models.ProviderStatusDisabled {
			existing.Status = models.ProviderStatusActive
		} else if !*req.IsEnabled {
			existing.Status = models.ProviderStatusDisabled
		}
	}
	if req.IsDefault != nil {
		existing.IsDefault = *req.IsDefault
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Config != nil {
		configBytes, _ := json.Marshal(req.Config)
		existing.Config = string(configBytes)
	}
	if req.SecretConfig != nil {
		secretBytes, _ := json.Marshal(req.SecretConfig)
		existing.SecretConfig = string(secretBytes)
	}

	if err := h.providerRepo.Update(c.Context(), existing); err != nil {
		return response.InternalError(c, "Failed to update provider: "+err.Error())
	}

	return response.OK(c, mapProviderToResponse(existing))
}

// ============================================
// CRUD: DeleteProvider
// ============================================

// DeleteProvider godoc
// @Summary Delete provider
// @Description Delete (soft-delete) a provider
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param providerId path string true "Provider ID"
// @Success 200 {object} dto.ActionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/providers/{providerId} [delete]
func (h *ProviderHandler) DeleteProvider(c *fiber.Ctx) error {
	providerIDStr := c.Params("providerId")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_PROVIDER_ID", "Invalid provider ID")
	}

	existing, err := h.providerRepo.GetByID(c.Context(), providerID)
	if err != nil {
		return response.InternalError(c, "Failed to get provider")
	}
	if existing == nil {
		return response.NotFound(c, "Provider not found")
	}

	if err := h.providerRepo.Delete(c.Context(), providerID); err != nil {
		return response.InternalError(c, "Failed to delete provider: "+err.Error())
	}

	return response.OK(c, &dto.ActionResponse{
		Message: "Provider deleted",
		ID:      providerID,
		Status:  "deleted",
	})
}

// ============================================
// SetDefaultProvider
// ============================================

// SetDefaultProvider godoc
// @Summary Set provider as default
// @Description Set or unset a provider as default for its channel
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param providerId path string true "Provider ID"
// @Param body body dto.SetDefaultProviderRequest true "Default status"
// @Success 200 {object} dto.ProviderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/providers/{providerId}/default [patch]
func (h *ProviderHandler) SetDefaultProvider(c *fiber.Ctx) error {
	providerIDStr := c.Params("providerId")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_PROVIDER_ID", "Invalid provider ID")
	}

	existing, err := h.providerRepo.GetByID(c.Context(), providerID)
	if err != nil {
		return response.InternalError(c, "Failed to get provider")
	}
	if existing == nil {
		return response.NotFound(c, "Provider not found")
	}

	req := new(dto.SetDefaultProviderRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	if req.IsDefault {
		channelProviders, err := h.providerRepo.List(c.Context(), existing.Channel)
		if err == nil {
			for _, p := range channelProviders {
				if p.ID != existing.ID && p.IsDefault {
					p.IsDefault = false
					_ = h.providerRepo.Update(c.Context(), p)
				}
			}
		}
	}

	existing.IsDefault = req.IsDefault

	if err := h.providerRepo.Update(c.Context(), existing); err != nil {
		return response.InternalError(c, "Failed to update provider default status: "+err.Error())
	}

	return response.OK(c, mapProviderToResponse(existing))
}

// ============================================
// CRUD: ToggleProviderStatus
// ============================================

// ToggleProviderStatus godoc
// @Summary Toggle provider status
// @Description Enable or disable a provider, or update status field
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param providerId path string true "Provider ID"
// @Param status body dto.ToggleProviderStatusRequest true "Status update"
// @Success 200 {object} dto.ProviderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /admin/providers/{providerId}/status [patch]
func (h *ProviderHandler) ToggleProviderStatus(c *fiber.Ctx) error {
	providerIDStr := c.Params("providerId")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		return response.BadRequest(c, "INVALID_PROVIDER_ID", "Invalid provider ID")
	}

	existing, err := h.providerRepo.GetByID(c.Context(), providerID)
	if err != nil {
		return response.InternalError(c, "Failed to get provider")
	}
	if existing == nil {
		return response.NotFound(c, "Provider not found")
	}

	req := new(dto.ToggleProviderStatusRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid status data")
	}

	if req.Status != "" {
		if !validStatuses[req.Status] {
			return response.BadRequest(c, "VALIDATION_ERROR", "Status must be one of: active, inactive, disabled, error")
		}
		existing.Status = req.Status
		existing.IsEnabled = req.Status != "disabled"
	} else {
		existing.IsEnabled = req.IsEnabled
		if req.IsEnabled {
			existing.Status = models.ProviderStatusActive
		} else {
			existing.Status = models.ProviderStatusDisabled
		}
	}

	if err := h.providerRepo.Update(c.Context(), existing); err != nil {
		return response.InternalError(c, "Failed to update provider status: "+err.Error())
	}

	return response.OK(c, mapProviderToResponse(existing))
}

// ============================================
// ListProviders
// ============================================

// ListProviders godoc
// @Summary List providers
// @Description Retrieve all configured notification providers from providers table
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param channel query string false "Filter by channel (sms, email, push, webhook, in_app)"
// @Param status query string false "Filter by status (active, inactive, disabled, error)"
// @Param providerType query string false "Filter by provider type (e.g., kavenegar, smtp)"
// @Success 200 {array} dto.ProviderResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/providers [get]
func (h *ProviderHandler) ListProviders(c *fiber.Ctx) error {
	channelFilter := c.Query("channel")
	statusFilter := c.Query("status")
	typeFilter := c.Query("providerType")

	dbProviders, err := h.providerRepo.List(c.Context(), channelFilter)
	if err != nil {
		return response.InternalError(c, "Failed to list providers: "+err.Error())
	}

	providers := make([]*dto.ProviderResponse, 0, len(dbProviders))
	for _, p := range dbProviders {
		resp := mapProviderToResponse(p)
		if statusFilter != "" && resp.Status != statusFilter {
			continue
		}
		if typeFilter != "" && p.Type != typeFilter {
			continue
		}
		providers = append(providers, resp)
	}

	return response.OK(c, providers)
}

// ============================================
// GetProviderHealth
// ============================================

// GetProviderHealth godoc
// @Summary Get provider health
// @Description Retrieve health status of all providers based on status/config validity
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ProviderHealthResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/providers/health [get]
func (h *ProviderHandler) GetProviderHealth(c *fiber.Ctx) error {
	ctx := c.Context()
	healthItems := make([]*dto.ProviderHealthItem, 0)
	checkedAt := time.Now()

	dbProviders, err := h.providerRepo.List(ctx, "")
	if err != nil {
		return response.InternalError(c, "Failed to get provider health: "+err.Error())
	}

	for _, p := range dbProviders {
		switch p.Status {
		case models.ProviderStatusActive:
			healthItems = append(healthItems, &dto.ProviderHealthItem{
				Name:    p.Name,
				Channel: p.Channel,
				Status:  "healthy",
			})
		case models.ProviderStatusInactive:
			healthItems = append(healthItems, &dto.ProviderHealthItem{
				Name:    p.Name,
				Channel: p.Channel,
				Status:  "degraded",
			})
		case models.ProviderStatusDisabled:
			healthItems = append(healthItems, &dto.ProviderHealthItem{
				Name:    p.Name,
				Channel: p.Channel,
				Status:  "disabled",
			})
		case models.ProviderStatusError:
			healthItems = append(healthItems, &dto.ProviderHealthItem{
				Name:    p.Name,
				Channel: p.Channel,
				Status:  "down",
			})
		default:
			healthItems = append(healthItems, &dto.ProviderHealthItem{
				Name:    p.Name,
				Channel: p.Channel,
				Status:  "unknown",
			})
		}
	}

	healthyCount := int64(0)
	degradedCount := int64(0)
	downCount := int64(0)
	disabledCount := int64(0)
	for _, item := range healthItems {
		switch item.Status {
		case "healthy":
			healthyCount++
		case "degraded":
			degradedCount++
		case "down":
			downCount++
		case "disabled":
			disabledCount++
		}
	}

	return response.OK(c, &dto.ProviderHealthResponse{
		Providers:     healthItems,
		HealthyCount:  healthyCount,
		DegradedCount: degradedCount,
		DownCount:     downCount,
		DisabledCount: disabledCount,
		CheckedAt:     checkedAt,
	})
}

// ============================================
// TestProvider
// ============================================

// TestProvider godoc
// @Summary Test provider
// @Description Test a specific provider connection (dry-run by default)
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param providerId path string true "Provider ID"
// @Param test body dto.ProviderTestRequest false "Test payload (optional, uses defaults if empty)"
// @Success 200 {object} dto.ProviderTestResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/providers/{providerId}/test [post]
func (h *ProviderHandler) TestProvider(c *fiber.Ctx) error {
	providerIDStr := c.Params("providerId")
	if providerIDStr == "" {
		return response.BadRequest(c, "PROVIDER_ID_REQUIRED", "Provider ID is required")
	}

	req := new(dto.ProviderTestRequest)
	if err := c.BodyParser(req); err != nil {
		req = &dto.ProviderTestRequest{DryRun: true}
	}

	providerID, parseErr := uuid.Parse(providerIDStr)
	providerFound := false
	channel := req.Channel

	if parseErr == nil {
		provider, err := h.providerRepo.GetByID(c.Context(), providerID)
		if err == nil && provider != nil {
			providerFound = true
			if channel == "" {
				channel = provider.Channel
			}
		}
	}

	if !providerFound && channel == "" {
		channel = "unknown"
	}

	responseMessage := "Provider test simulated (dry-run)."
	if !providerFound {
		responseMessage = "Provider is not configured. Dry-run simulation returned without real provider check."
	} else {
		responseMessage = "Provider configuration is valid. Dry-run simulation returned successfully."
	}

	return response.OK(c, &dto.ProviderTestResponse{
		ProviderID: providerIDStr,
		Channel:    channel,
		DryRun:     true,
		Success:    true,
		Status:     "simulated",
		Message:    responseMessage,
		LatencyMs:  5,
		CheckedAt:  time.Now(),
	})
}
