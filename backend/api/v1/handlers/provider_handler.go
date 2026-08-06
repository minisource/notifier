package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/provider"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/service"

)

// ProviderHandler handles provider-related admin endpoints
type ProviderHandler struct {
	providerRepo    repository.ProviderRepository
	notifRepo       repository.NotificationRepository
	logger          logging.Logger
	deliveryControl *service.DeliveryControlService
}

// NewProviderHandler creates a new provider handler. deliveryControl wires the
// global outbound delivery pause into the provider test-send path so a real
// (non dry-run) test message is never sent while delivery is frozen.
func NewProviderHandler(
	providerRepo repository.ProviderRepository,
	notifRepo repository.NotificationRepository,
	logger logging.Logger,
	deliveryControl *service.DeliveryControlService,
) *ProviderHandler {
	return &ProviderHandler{
		providerRepo:    providerRepo,
		notifRepo:       notifRepo,
		logger:          logger,
		deliveryControl: deliveryControl,
	}
}

// providerStatsMap loads per-provider delivery stats keyed by provider name/type
// so the UI can show success rate, latency, and last failure for each provider.
func (h *ProviderHandler) providerStatsMap(ctx context.Context) map[string]*repository.ProviderStats {
	stats, err := h.notifRepo.GetProviderStats(ctx)
	if err != nil {
		return map[string]*repository.ProviderStats{}
	}
	m := make(map[string]*repository.ProviderStats, len(stats))
	for _, s := range stats {
		m[s.Provider] = s
	}
	return m
}

// applyProviderStats merges aggregated delivery stats into a provider response.
// Stats are keyed by the provider name stored on notifications — fall back to
// the provider type when the name has no history yet.
func applyProviderStats(resp *dto.ProviderResponse, p *models.Provider, stats map[string]*repository.ProviderStats) {
	if resp == nil {
		return
	}
	// The adapters record config.Provider (which is the provider Type, e.g.
	// 'kavenegar') on notifications — so look up by type first, fall back to
	// the display name for legacy data.
	stat := stats[p.Type]
	if stat == nil {
		stat = stats[p.Name]
	}
	if stat == nil {
		return
	}
	resp.SuccessRate = stat.SuccessRate
	resp.AverageLatencyMs = int64(stat.AverageLatencyMs)
	resp.LastSuccessAt = stat.LastSuccessAt
	resp.LastFailureAt = stat.LastFailureAt
	resp.LastError = stat.LastError
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

// resolveTenantID returns the effective tenant scope for this request.
// Priority: tenantId query param > X-Tenant-Id header. Returns nil when the
// caller is in the global/all-tenants scope (no tenant filter applied).
func resolveTenantID(c *fiber.Ctx) *uuid.UUID {
	raw := c.Query("tenantId")
	if raw == "" {
		raw = c.Get("X-Tenant-Id")
	}
	if raw == "" || raw == "all" || raw == "default" {
		return nil
	}
	if tid, err := uuid.Parse(raw); err == nil {
		return &tid
	}
	return nil
}

// providerAccessible reports whether a provider is visible/manageable under the
// given tenant scope. Global providers (tenant_id NULL) are visible to every
// tenant; a tenant-specific provider is only visible inside its own tenant.
// A nil scope (global view) can see everything.
func providerAccessible(p *models.Provider, tenantID *uuid.UUID) bool {
	if tenantID == nil {
		return true
	}
	if p.TenantID == nil {
		return true
	}
	return *p.TenantID == *tenantID
}

// sameTenantScope reports whether two tenant scopes are exactly equal,
// treating nil (global) as a distinct scope from any specific tenant.
func sameTenantScope(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
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
		TenantID:    p.TenantID,
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

	// Tenant scoping: an explicit body tenantId is authoritative. "all"/"default"
	// means global (nil); a UUID means a specific tenant. Only when the body
	// omits tenantId entirely do we fall back to the X-Tenant-Id header so
	// providers created while a tenant is active belong to that tenant.
	var tenantID *uuid.UUID
	if req.TenantID != "" {
		if req.TenantID == "all" || req.TenantID == "default" {
			tenantID = nil
		} else {
			tid, err := uuid.Parse(req.TenantID)
			if err != nil {
				return response.BadRequest(c, "INVALID_TENANT_ID", "tenantId is not a valid tenant UUID")
			}
			tenantID = &tid
		}
	} else {
		tenantID = resolveTenantID(c)
	}

	provider := &models.Provider{
		TenantID:     tenantID,
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

	resp := mapProviderToResponse(provider)
	applyProviderStats(resp, provider, h.providerStatsMap(c.Context()))
	return response.Created(c, resp)
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

	// Tenant isolation: do not leak providers from other tenants
	if !providerAccessible(provider, resolveTenantID(c)) {
		return response.NotFound(c, "Provider not found")
	}

	resp := mapProviderToResponse(provider)
	applyProviderStats(resp, provider, h.providerStatsMap(c.Context()))
	return response.OK(c, resp)
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

	// Tenant isolation
	if !providerAccessible(existing, resolveTenantID(c)) {
		return response.NotFound(c, "Provider not found")
	}

	req := new(dto.UpdateProviderRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid provider data")
	}

	// Tenant re-scoping: "all"/"default" clears to global, a UUID moves the
	// provider to that tenant, empty leaves the current tenant unchanged.
	if req.TenantID != "" {
		if req.TenantID == "all" || req.TenantID == "default" {
			existing.TenantID = nil
		} else {
			tid, err := uuid.Parse(req.TenantID)
			if err != nil {
				return response.BadRequest(c, "INVALID_TENANT_ID", "tenantId is not a valid tenant UUID")
			}
			existing.TenantID = &tid
		}
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

	resp := mapProviderToResponse(existing)
	applyProviderStats(resp, existing, h.providerStatsMap(c.Context()))
	return response.OK(c, resp)
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

	// Tenant isolation
	if !providerAccessible(existing, resolveTenantID(c)) {
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

	// Tenant isolation
	if !providerAccessible(existing, resolveTenantID(c)) {
		return response.NotFound(c, "Provider not found")
	}

	req := new(dto.SetDefaultProviderRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	if req.IsDefault {
		// Only unset defaults within the exact same tenant scope: a global default
		// must not clobber tenant-specific defaults and vice versa.
		channelProviders, err := h.providerRepo.List(c.Context(), existing.Channel, existing.TenantID)
		if err == nil {
			for _, p := range channelProviders {
				if p.ID != existing.ID && sameTenantScope(p.TenantID, existing.TenantID) && p.IsDefault {
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

	resp := mapProviderToResponse(existing)
	applyProviderStats(resp, existing, h.providerStatsMap(c.Context()))
	return response.OK(c, resp)
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

	// Tenant isolation
	if !providerAccessible(existing, resolveTenantID(c)) {
		return response.NotFound(c, "Provider not found")
	}

	req := new(dto.ToggleProviderStatusRequest)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_STATUS", "Invalid status data")
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

	resp := mapProviderToResponse(existing)
	applyProviderStats(resp, existing, h.providerStatsMap(c.Context()))
	return response.OK(c, resp)
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
	tenantID := resolveTenantID(c)

	dbProviders, err := h.providerRepo.List(c.Context(), channelFilter, tenantID)
	if err != nil {
		return response.InternalError(c, "Failed to list providers: "+err.Error())
	}

	stats := h.providerStatsMap(c.Context())
	providers := make([]*dto.ProviderResponse, 0, len(dbProviders))
	for _, p := range dbProviders {
		resp := mapProviderToResponse(p)
		applyProviderStats(resp, p, stats)
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
// providerHealthCheck — REAL checks against each provider's own API
// ============================================

// checkAllProviders runs a REAL connectivity/credential check against every
// enabled provider's own API (Kavenegar account/info, SMTP handshake, ...).
// Each check is bounded by its own timeout so one hung provider can't stall
// the rest. Disabled providers are reported without hitting the network.
func (h *ProviderHandler) checkAllProviders(c *fiber.Ctx) *dto.ProviderHealthResponse {
	ctx := c.Context()
	checkedAt := time.Now()

	dbProviders, err := h.providerRepo.List(ctx, "", resolveTenantID(c))
	if err != nil {
		h.logger.Warn(logging.General, logging.Select, "Failed to list providers for health check", map[logging.ExtraKey]interface{}{
			logging.ExtraKey("error"): err.Error(),
		})
		return &dto.ProviderHealthResponse{CheckedAt: checkedAt}
	}

	type itemWithIdx struct {
		idx  int
		item *dto.ProviderHealthItem
	}

	items := make([]*dto.ProviderHealthItem, len(dbProviders))
	results := make(chan itemWithIdx, len(dbProviders))
	var wg sync.WaitGroup

	for i, p := range dbProviders {
		wg.Add(1)
		go func(idx int, prov *models.Provider) {
			defer wg.Done()
			checkCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
			defer cancel()
			res := provider.CheckProvider(checkCtx, prov)
			results <- itemWithIdx{idx: idx, item: &dto.ProviderHealthItem{
				ProviderID: res.ProviderID,
				Name:       res.Name,
				Channel:    res.Channel,
				Type:       res.Type,
				Status:     res.Status,
				LatencyMs:  res.LatencyMs,
				Message:    res.Message,
				Error:      res.Error,
				CheckedAt:  res.CheckedAt,
			}}
		}(i, p)
	}
	wg.Wait()
	close(results)

	for r := range results {
		items[r.idx] = r.item
	}

	var healthyCount, degradedCount, downCount, disabledCount int64
	for _, item := range items {
		if item == nil {
			continue
		}
		switch item.Status {
		case "healthy":
			healthyCount++
		case "degraded", "unsupported":
			degradedCount++
		case "down":
			downCount++
		case "disabled":
			disabledCount++
		}
	}

	return &dto.ProviderHealthResponse{
		Providers:     items,
		HealthyCount:  healthyCount,
		DegradedCount: degradedCount,
		DownCount:     downCount,
		DisabledCount: disabledCount,
		CheckedAt:     checkedAt,
	}
}

// ============================================
// GetProviderHealth
// ============================================

// GetProviderHealth godoc
// @Summary Get provider health
// @Description Run real connectivity checks against every provider's own API and return health status
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ProviderHealthResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/providers/health [get]
func (h *ProviderHandler) GetProviderHealth(c *fiber.Ctx) error {
	return response.OK(c, h.checkAllProviders(c))
}

// ============================================
// HealthCheckAllProviders
// ============================================

// HealthCheckAllProviders godoc
// @Summary Check all providers health
// @Description Run real connectivity checks against all providers' own APIs and return the aggregate result
// @Tags Providers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ProviderHealthResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/providers/health-check [post]
func (h *ProviderHandler) HealthCheckAllProviders(c *fiber.Ctx) error {
	return response.OK(c, h.checkAllProviders(c))
}

// ============================================
// TestProvider
// ============================================

// TestProvider godoc
// @Summary Test provider
// @Description Run a real connectivity check against the provider's own API (dry-run), or send a real test message when dryRun=false and a recipient is provided
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
		req = &dto.ProviderTestRequest{}
	}

	// DryRun defaults to true when omitted: a plain connectivity/credential check.
	dryRun := true
	if req.DryRun != nil {
		dryRun = *req.DryRun
	}

	providerID, parseErr := uuid.Parse(providerIDStr)
	if parseErr != nil {
		return response.BadRequest(c, "INVALID_PROVIDER_ID", "Invalid provider ID")
	}

	prov, err := h.providerRepo.GetByID(c.Context(), providerID)
	if err != nil {
		return response.InternalError(c, "Failed to get provider")
	}
	if prov == nil {
		return response.NotFound(c, "Provider not found")
	}
	if !providerAccessible(prov, resolveTenantID(c)) {
		return response.NotFound(c, "Provider not found")
	}

	channel := prov.Channel

	// ---- Non dry-run: send a REAL test message through the provider ----
	// A test-send is an outbound delivery — it is frozen while the global
	// outbound delivery pause is active, exactly like retries/fallbacks/manual
	// resends. The backend gate is authoritative regardless of UI state.
	if !dryRun && h.deliveryControl != nil && h.deliveryControl.IsPaused(c.Context()) {
		return response.BadRequest(c, "DELIVERY_PAUSED", "Outbound delivery is paused — test messages cannot be sent until delivery is resumed")
	}

	if !dryRun {
		if req.Recipient == "" {
			return response.BadRequest(c, "RECIPIENT_REQUIRED", "recipient is required for a real (non dry-run) provider test")
		}

		testCtx, cancel := context.WithTimeout(c.Context(), 20*time.Second)
		defer cancel()

		msgID, latency, sendErr := provider.SendTestMessage(testCtx, prov, req.Recipient, req.Subject, req.Body)
		if sendErr != nil {
			return response.OK(c, &dto.ProviderTestResponse{
				ProviderID: providerIDStr,
				Channel:    channel,
				DryRun:     false,
				Success:    false,
				Status:     "failed",
				Message:    "Provider send failed: " + sendErr.Error(),
				LatencyMs:  latency,
				CheckedAt:  time.Now(),
			})
		}

		return response.OK(c, &dto.ProviderTestResponse{
			ProviderID:                providerIDStr,
			Channel:                   channel,
			DryRun:                    false,
			Success:                   true,
			Status:                    "sent",
			Message:                   "Provider accepted the test message.",
			ProviderMessageID:         msgID,
			ProviderResponseSanitized: "accepted",
			LatencyMs:                 latency,
			CheckedAt:                 time.Now(),
		})
	}

	// ---- Dry-run (default): REAL connectivity/credential check ----
	checkCtx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
	defer cancel()

	res := provider.CheckProvider(checkCtx, prov)

	if res.Status == "healthy" {
		return response.OK(c, &dto.ProviderTestResponse{
			ProviderID:                providerIDStr,
			Channel:                   channel,
			DryRun:                    true,
			Success:                   true,
			Status:                    "healthy",
			Message:                   res.Message,
			ProviderResponseSanitized: res.Message,
			LatencyMs:                 res.LatencyMs,
			CheckedAt:                 res.CheckedAt,
		})
	}

	return response.OK(c, &dto.ProviderTestResponse{
		ProviderID:                providerIDStr,
		Channel:                   channel,
		DryRun:                    true,
		Success:                   false,
		Status:                    res.Status,
		Message:                   res.Message,
		ProviderResponseSanitized: res.Error,
		LatencyMs:                 res.LatencyMs,
		CheckedAt:                 res.CheckedAt,
	})
}
