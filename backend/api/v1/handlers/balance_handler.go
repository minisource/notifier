package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/service"
)

// BalanceHandler exposes provider account balance/quota health, history, alerts,
// and monitoring settings to authorized admins. All values are already
// sanitized by the service/provider layer — no secrets can reach this handler.
type BalanceHandler struct {
	balanceService *service.BalanceService
	providerRepo   repository.ProviderRepository
	balanceRepo    repository.ProviderBalanceRepository
	logger         logging.Logger
}

// NewBalanceHandler creates the handler.
func NewBalanceHandler(
	balanceService *service.BalanceService,
	providerRepo repository.ProviderRepository,
	balanceRepo repository.ProviderBalanceRepository,
	logger logging.Logger,
) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
		providerRepo:   providerRepo,
		balanceRepo:    balanceRepo,
		logger:         logger,
	}
}

// ListHealth godoc
// @Summary List provider account health
// @Description Return the current balance/quota health of every provider account (persisted values, no external calls)
// @Tags ProviderBalance
// @Security BearerAuth
// @Success 200 {array} dto.ProviderAccountHealthSummary
// @Router /admin/providers/balance [get]
func (h *BalanceHandler) ListHealth(c *fiber.Ctx) error {
	healthRows, err := h.balanceRepo.ListHealth(c.Context(), resolveTenantID(c))
	if err != nil {
		return response.InternalError(c, "Failed to list provider health")
	}
	out := make([]*dto.ProviderAccountHealthSummary, 0, len(healthRows))
	for _, hRow := range healthRows {
		out = append(out, mapHealthSummary(hRow))
	}
	return response.OK(c, out)
}

// GetHealth godoc
// @Summary Get provider account health + history
// @Description Return current health, balance history, and alerts for one provider account
// @Tags ProviderBalance
// @Security BearerAuth
// @Param providerId path string true "Provider ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/providers/{providerId}/balance [get]
func (h *BalanceHandler) GetHealth(c *fiber.Ctx) error {
	providerID, ok := parseProviderID(c)
	if !ok {
		return response.BadRequest(c, "INVALID_PROVIDER_ID", "Invalid provider ID")
	}
	p, err := h.providerRepo.GetByID(c.Context(), providerID)
	if err != nil || p == nil {
		return response.NotFound(c, "Provider not found")
	}
	if !providerAccessible(p, resolveTenantID(c)) {
		return response.NotFound(c, "Provider not found")
	}

	health, _ := h.balanceRepo.GetHealth(c.Context(), providerID)
	history, _ := h.balanceRepo.ListSnapshots(c.Context(), providerID, nil, nil, 100)
	alerts, _ := h.balanceRepo.ListAlerts(c.Context(), providerID, "", 50)

	summary := mapHealthSummary(health)
	activeCount := 0
	highest := ""
	for _, a := range alerts {
		if a.Status == models.AlertStatusActive {
			activeCount++
			if severityRank(a.Severity) > severityRank(highest) {
				highest = a.Severity
			}
		}
	}
	if summary != nil {
		summary.ActiveAlertCount = activeCount
		if summary.LatestAlertLevel == "" {
			summary.LatestAlertLevel = highest
		}
	}

	return response.OK(c, fiber.Map{
		"providerId": providerID.String(),
		"name":       p.Name,
		"channel":    p.Channel,
		"type":       p.Type,
		"health":     summary,
		"history":    mapSnapshots(history),
		"alerts":     mapAlerts(alerts),
		"settings":   mapSettings(service.ParseBalanceSettings(providerConfigJSONForHandler(p))),
	})
}

// Refresh godoc
// @Summary Manually refresh a provider account balance
// @Description Trigger a real balance fetch now
// @Tags ProviderBalance
// @Security BearerAuth
// @Param providerId path string true "Provider ID"
// @Success 200 {object} dto.BalanceRefreshResponse
// @Router /admin/providers/{providerId}/balance/refresh [post]
func (h *BalanceHandler) Refresh(c *fiber.Ctx) error {
	providerID, ok := parseProviderID(c)
	if !ok {
		return response.BadRequest(c, "INVALID_PROVIDER_ID", "Invalid provider ID")
	}
	p, err := h.providerRepo.GetByID(c.Context(), providerID)
	if err != nil || p == nil {
		return response.NotFound(c, "Provider not found")
	}
	if !providerAccessible(p, resolveTenantID(c)) {
		return response.NotFound(c, "Provider not found")
	}

	res, err := h.balanceService.RefreshAccount(c.Context(), providerID)
	if err != nil {
		return response.InternalError(c, "Balance refresh failed: "+err.Error())
	}
	return response.OK(c, &dto.BalanceRefreshResponse{
		ProviderID:     res.ProviderID,
		Name:           res.Name,
		Channel:        res.Channel,
		CapabilityMode: res.CapabilityMode,
		Success:        res.Success,
		HealthLevel:    res.HealthLevel,
		BalanceValue:   res.BalanceValue,
		BalanceUnit:    res.BalanceUnit,
		Currency:       res.Currency,
		ErrorKind:      res.ErrorKind,
		ErrorCode:      res.ErrorCode,
		ErrorMessage:   res.ErrorMessage,
		LatencyMs:      res.LatencyMs,
		CheckedAt:      res.CheckedAt,
	})
}

// UpdateSettings godoc
// @Summary Update per-account balance monitoring settings
// @Description Persist thresholds (in provider unit) and monitoring toggle into the provider config (balanceSettings)
// @Tags ProviderBalance
// @Security BearerAuth
// @Param providerId path string true "Provider ID"
// @Param body body dto.BalanceSettingsDTO true "Settings"
// @Success 200 {object} dto.BalanceSettingsDTO
// @Router /admin/providers/{providerId}/balance/settings [put]
func (h *BalanceHandler) UpdateSettings(c *fiber.Ctx) error {
	providerID, ok := parseProviderID(c)
	if !ok {
		return response.BadRequest(c, "INVALID_PROVIDER_ID", "Invalid provider ID")
	}
	p, err := h.providerRepo.GetByID(c.Context(), providerID)
	if err != nil || p == nil {
		return response.NotFound(c, "Provider not found")
	}
	if !providerAccessible(p, resolveTenantID(c)) {
		return response.NotFound(c, "Provider not found")
	}

	req := new(dto.BalanceSettingsDTO)
	if err := c.BodyParser(req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid settings payload")
	}
	// Validate thresholds: critical must not be higher than warning.
	if req.WarningThreshold != nil && req.CriticalThreshold != nil && *req.CriticalThreshold > *req.WarningThreshold {
		return response.BadRequest(c, "VALIDATION_ERROR", "criticalThreshold must be lower than warningThreshold")
	}
	if req.RefreshIntervalSec != nil && *req.RefreshIntervalSec < 60 {
		return response.BadRequest(c, "VALIDATION_ERROR", "refreshIntervalSec must be at least 60")
	}

	if err := h.saveBalanceSettings(c, p, req); err != nil {
		return response.InternalError(c, "Failed to save balance settings: "+err.Error())
	}
	return response.OK(c, req)
}

// ListAlerts godoc
// @Summary List credit alerts
// @Description Return alert occurrences (optionally filtered by status), newest first
// @Tags ProviderBalance
// @Security BearerAuth
// @Param status query string false "active | acknowledged | resolved"
// @Param providerId query string false "Filter by provider"
// @Success 200 {array} dto.CreditAlertDTO
// @Router /admin/providers/balance/alerts [get]
func (h *BalanceHandler) ListAlerts(c *fiber.Ctx) error {
	status := c.Query("status")
	pid := c.Query("providerId")

	if pid != "" {
		providerID, err := uuid.Parse(pid)
		if err != nil {
			return response.BadRequest(c, "INVALID_PROVIDER_ID", "Invalid provider ID")
		}
		rows, err := h.balanceRepo.ListAlerts(c.Context(), providerID, status, 100)
		if err != nil {
			return response.InternalError(c, "Failed to list alerts")
		}
		return response.OK(c, mapAlerts(rows))
	}

	rows, err := h.balanceRepo.ListAllAlerts(c.Context(), status, resolveTenantID(c), 100)
	if err != nil {
		return response.InternalError(c, "Failed to list alerts")
	}
	return response.OK(c, mapAlerts(rows))
}

// Acknowledge godoc
// @Summary Acknowledge a credit alert
// @Description Mark an alert acknowledged (does not resolve the underlying condition)
// @Tags ProviderBalance
// @Security BearerAuth
// @Param alertId path string true "Alert ID"
// @Success 200 {object} dto.CreditAlertDTO
// @Router /admin/providers/balance/alerts/{alertId}/acknowledge [post]
func (h *BalanceHandler) Acknowledge(c *fiber.Ctx) error {
	alertID, err := uuid.Parse(c.Params("alertId"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ALERT_ID", "Invalid alert ID")
	}
	alert, err := h.getAlertByID(c, alertID)
	if err != nil || alert == nil {
		return response.NotFound(c, "Alert not found")
	}
	now := time.Now().UTC()
	alert.Status = models.AlertStatusAcknowledged
	alert.AcknowledgedAt = &now
	userID := c.Locals("userId")
	if uid, ok := userID.(string); ok && uid != "" {
		alert.AcknowledgedBy = uid
	} else {
		alert.AcknowledgedBy = "admin"
	}
	if err := h.balanceRepo.UpdateAlert(c.Context(), alert); err != nil {
		return response.InternalError(c, "Failed to acknowledge alert")
	}
	return response.OK(c, mapAlert(alert))
}

// ---- helpers ----

func parseProviderID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Params("providerId"))
	return id, err == nil
}

func severityRank(s string) int {
	switch s {
	case "exhausted":
		return 4
	case "critical":
		return 3
	case "warning":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}

func mapHealthSummary(h *models.ProviderAccountHealth) *dto.ProviderAccountHealthSummary {
	if h == nil {
		return nil
	}
	return &dto.ProviderAccountHealthSummary{
		ProviderID:              h.ProviderID.String(),
		TenantID:                h.TenantID,
		Provider:                h.Provider,
		Channel:                 h.Channel,
		CapabilityMode:          h.CapabilityMode,
		HealthLevel:             h.HealthLevel,
		LatestAlertLevel:        h.LatestAlertLevel,
		BalanceValue:            h.BalanceValue,
		BalanceUnit:             h.BalanceUnit,
		Currency:                h.Currency,
		QuotaRemaining:          h.QuotaRemaining,
		QuotaLimit:              h.QuotaLimit,
		UsagePercent:            h.UsagePercent,
		IsEstimated:             h.IsEstimated,
		IsManual:                h.IsManual,
		Source:                  h.Source,
		LastSuccessfulRefreshAt: h.LastSuccessfulRefreshAt,
		LastRefreshAttemptAt:    h.LastRefreshAttemptAt,
		NextScheduledRefreshAt:  h.NextScheduledRefreshAt,
		ConsecutiveFailures:     h.ConsecutiveFailures,
		LastErrorKind:           h.LastErrorKind,
		LastErrorMessage:        h.LastErrorMessage,
		UpdatedAt:               h.UpdatedAt,
	}
}

func mapSnapshots(rows []*models.ProviderBalanceSnapshot) []*dto.BalanceSnapshotDTO {
	out := make([]*dto.BalanceSnapshotDTO, 0, len(rows))
	for _, s := range rows {
		out = append(out, &dto.BalanceSnapshotDTO{
			ID:             s.ID.String(),
			ProviderID:     s.ProviderID.String(),
			RefreshStatus:  s.RefreshStatus,
			CapabilityMode: s.CapabilityMode,
			Source:         s.Source,
			IsEstimated:    s.IsEstimated,
			IsManual:       s.IsManual,
			BalanceValue:   s.BalanceValue,
			BalanceUnit:    s.BalanceUnit,
			Currency:       s.Currency,
			QuotaRemaining: s.QuotaRemaining,
			QuotaLimit:     s.QuotaLimit,
			UsagePercent:   s.UsagePercent,
			AccountStatus:  s.AccountStatus,
			PlanExpiresAt:  s.PlanExpiresAt,
			ErrorKind:      s.ErrorKind,
			ErrorCode:      s.ErrorCode,
			ErrorMessage:   s.ErrorMessage,
			FetchedAt:      s.FetchedAt,
			LatencyMs:      s.LatencyMs,
		})
	}
	return out
}

func mapAlert(a *models.ProviderCreditAlert) *dto.CreditAlertDTO {
	if a == nil {
		return nil
	}
	return &dto.CreditAlertDTO{
		ID:               a.ID.String(),
		ProviderID:       a.ProviderID.String(),
		Provider:         a.Provider,
		AlertType:        a.AlertType,
		Severity:         a.Severity,
		Status:           a.Status,
		Message:          a.Message,
		BalanceValue:     a.BalanceValue,
		ThresholdValue:   a.ThresholdValue,
		FirstTriggeredAt: a.FirstTriggeredAt,
		LastTriggeredAt:  a.LastTriggeredAt,
		RepeatCount:      a.RepeatCount,
		AcknowledgedAt:   a.AcknowledgedAt,
		AcknowledgedBy:   a.AcknowledgedBy,
		ResolvedAt:       a.ResolvedAt,
		ResolvedReason:   a.ResolvedReason,
	}
}

func mapAlerts(rows []*models.ProviderCreditAlert) []*dto.CreditAlertDTO {
	out := make([]*dto.CreditAlertDTO, 0, len(rows))
	for _, a := range rows {
		out = append(out, mapAlert(a))
	}
	return out
}

func mapSettings(s service.BalanceSettings) *dto.BalanceSettingsDTO {
	return &dto.BalanceSettingsDTO{
		Enabled:            s.Enabled,
		WarningThreshold:   s.WarningThreshold,
		CriticalThreshold:  s.CriticalThreshold,
		RefreshIntervalSec: s.RefreshIntervalSec,
	}
}

// getAlertByID fetches a single alert across all providers (admin endpoint).
func (h *BalanceHandler) getAlertByID(c *fiber.Ctx, alertID uuid.UUID) (*models.ProviderCreditAlert, error) {
	rows, err := h.balanceRepo.ListAllAlerts(c.Context(), "", nil, 2000)
	if err != nil {
		return nil, err
	}
	for _, a := range rows {
		if a.ID == alertID {
			return a, nil
		}
	}
	return nil, nil
}

// providerConfigJSONForHandler returns the public config JSON of a provider.
func providerConfigJSONForHandler(p *models.Provider) string {
	if p == nil {
		return ""
	}
	return p.Config
}

// saveBalanceSettings merges balanceSettings into the provider's public Config
// JSON and persists it through the handler's provider repository.
func (h *BalanceHandler) saveBalanceSettings(c *fiber.Ctx, p *models.Provider, req *dto.BalanceSettingsDTO) error {
	cfg := map[string]interface{}{}
	if p.Config != "" {
		if err := json.Unmarshal([]byte(p.Config), &cfg); err != nil {
			return err
		}
	}
	cfg["balanceSettings"] = map[string]interface{}{
		"enabled":            req.Enabled,
		"warningThreshold":   req.WarningThreshold,
		"criticalThreshold":  req.CriticalThreshold,
		"refreshIntervalSec": req.RefreshIntervalSec,
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	p.Config = string(b)
	return h.providerRepo.Update(c.Context(), p)
}
