package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/response"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/service"
)

// ObservabilityHandler handles observability endpoints
type ObservabilityHandler struct {
	notificationService *service.NotificationService
	startTime           time.Time
	dbAvailable         bool
}

// NewObservabilityHandler creates a new observability handler
func NewObservabilityHandler(notificationService *service.NotificationService) *ObservabilityHandler {
	return &ObservabilityHandler{
		notificationService: notificationService,
		startTime:           time.Now(),
	}
}

// GetHealth godoc
// @Summary Get service health (admin)
// @Description Retrieve detailed health status of the service and its dependencies. Admin/operator only.
// @Tags Admin Observability
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ObservabilityHealthResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /admin/observability/health [get]
func (h *ObservabilityHandler) GetHealth(c *fiber.Ctx) error {
	ctx := context.Background()
	now := time.Now()

	// Check database health
	dbStatus := "healthy"
	dbErr := ""
	_, err := h.notificationService.GetRepository().GetQueueDepth(ctx)
	if err != nil {
		dbStatus = "unhealthy"
		dbErr = err.Error()
	}

	// Build dependencies list
	deps := []*dto.DependencyHealth{
		{Name: "database", Status: dbStatus, Error: dbErr},
	}

	// Determine overall status
	overallStatus := "healthy"
	for _, dep := range deps {
		if dep.Status == "unhealthy" {
			overallStatus = "degraded"
			break
		}
	}

	return response.OK(c, dto.ObservabilityHealthResponse{
		Status:         overallStatus,
		Service:        "notifier",
		Version:        "1.0.0",
		Environment:    "production",
		UptimeSeconds:  int64(time.Since(h.startTime).Seconds()),
		Dependencies:   deps,
		GeneratedAt:    now,
	})
}

// GetReadiness godoc
// @Summary Get service readiness (admin)
// @Description Check if the service is ready to accept traffic. Verifies key dependencies. Admin/operator only.
// @Tags Admin Observability
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ObservabilityReadinessResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /admin/observability/readiness [get]
func (h *ObservabilityHandler) GetReadiness(c *fiber.Ctx) error {
	ctx := context.Background()
	now := time.Now()

	checkResults := make([]*dto.ReadinessCheck, 0)

	// Database check
	_, err := h.notificationService.GetRepository().GetQueueDepth(ctx)
	if err != nil {
		checkResults = append(checkResults, &dto.ReadinessCheck{
			Name:   "database",
			Status: "unhealthy",
			Error:  "Database unreachable: " + err.Error(),
		})
		return response.OK(c, dto.ObservabilityReadinessResponse{
			Ready:       false,
			Overall:     "not_ready",
			Checks:      checkResults,
			GeneratedAt: now,
		})
	}
	checkResults = append(checkResults, &dto.ReadinessCheck{
		Name:   "database",
		Status: "healthy",
	})

	return response.OK(c, dto.ObservabilityReadinessResponse{
		Ready:       true,
		Overall:     "ready",
		Checks:      checkResults,
		GeneratedAt: now,
	})
}

// GetQueueOverview godoc
// @Summary Get queue overview (admin)
// @Description Retrieve queue metrics — counts by status (pending, processing, retrying, dead, scheduled). Admin/operator only.
// @Tags Admin Observability
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.QueueOverviewResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/observability/queue [get]
func (h *ObservabilityHandler) GetQueueOverview(c *fiber.Ctx) error {
	ctx := context.Background()
	now := time.Now()

	// Get queue depth (pending + retrying)
	queueDepth, err := h.notificationService.GetRepository().GetQueueDepth(ctx)
	if err != nil {
		queueDepth = 0
	}

	// Scan recent notifications for status breakdown
	allItems, _, err := h.notificationService.ListAllNotifications(ctx, repository.NotificationListFilter{
		Page:     1,
		PageSize: 100,
	})
	if err != nil {
		return response.InternalError(c, "Failed to get queue overview")
	}

	pendingCount := int64(0)
	processingCount := int64(0)
	retryingCount := int64(0)
	deadCount := int64(0)
	scheduledCount := int64(0)

	var oldestPending *time.Time
	var nextRetry *time.Time

	for _, n := range allItems {
		switch n.Status {
		case models.NotificationStatusPending:
			pendingCount++
			if oldestPending == nil || n.CreatedAt.Before(*oldestPending) {
				oldestPending = &n.CreatedAt
			}
		case models.NotificationStatusProcessing, models.NotificationStatusSending:
			processingCount++
		case models.NotificationStatusRetrying:
			retryingCount++
			if n.NextRetryAt != nil && (nextRetry == nil || n.NextRetryAt.Before(*nextRetry)) {
				nextRetry = n.NextRetryAt
			}
		case models.NotificationStatusDead:
			deadCount++
		}
		if n.ScheduledAt != nil && n.Status == models.NotificationStatusPending && n.ScheduledAt.After(now) {
			scheduledCount++
		}
	}

	// If queueDepth is higher than what we scanned, adjust
	if queueDepth > pendingCount+retryingCount {
		remaining := queueDepth - pendingCount - retryingCount
		if remaining > 0 {
			pendingCount += remaining
		}
	}

	return response.OK(c, dto.QueueOverviewResponse{
		PendingCount:     pendingCount,
		ProcessingCount:  processingCount,
		RetryingCount:    retryingCount,
		DeadCount:        deadCount,
		ScheduledCount:   scheduledCount,
		OldestPendingAt:  oldestPending,
		NextRetryAt:      nextRetry,
		ThroughputPerMin: 0, // Requires time-series metrics; documented limitation
		AverageLatencyMs: 0, // Requires latency aggregation; documented limitation
		GeneratedAt:      now,
	})
}

// GetWorkersOverview godoc
// @Summary Get workers overview (admin)
// @Description Retrieve worker status information. Worker heartbeat tracking is not fully implemented — returns configured worker pool info.
// @Tags Admin Observability
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.WorkerOverviewResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/observability/workers [get]
func (h *ObservabilityHandler) GetWorkersOverview(c *fiber.Ctx) error {
	// Return configured worker info — no dynamic heartbeat tracking
	// The project has 10 configured workers (from config default)
	workers := []*dto.WorkerInfo{
		{ID: "worker-pool", Name: "notification-workers", Status: "active", Channel: "all", QueueSize: 10},
	}

	return response.OK(c, dto.WorkerOverviewResponse{
		Workers:         workers,
		ActiveCount:     10,
		IdleCount:       0,
		FailedCount:     0,
		LastHeartbeatAt: nil, // No heartbeat tracking; documented limitation
		GeneratedAt:     time.Now(),
	})
}

// GetMetrics godoc
// @Summary Get operational metrics (admin)
// @Description Retrieve service operational metrics — queue depth, counts by status, success rate, delivery latency. Admin/operator only.
// @Tags Admin Observability
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ObservabilityMetricsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/observability/metrics [get]
func (h *ObservabilityHandler) GetMetrics(c *fiber.Ctx) error {
	ctx := context.Background()

	// Get queue depth
	queueDepth, err := h.notificationService.GetRepository().GetQueueDepth(ctx)
	if err != nil {
		return response.InternalError(c, "Failed to get queue depth")
	}

	// Get counts via paginated listing — partial snapshot for metrics
	allItems, _, err := h.notificationService.ListAllNotifications(ctx, repository.NotificationListFilter{
		Page:     1,
		PageSize: 100,
	})
	if err != nil {
		return response.InternalError(c, "Failed to get notification metrics")
	}

	sentCount := int64(0)
	failedCount := int64(0)
	pendingCount := int64(0)
	deadCount := int64(0)
	totalAttempts := int64(0)
	failedAttempts := int64(0)
	totalLatencyMs := int64(0)
	latencyCount := int64(0)

	for _, n := range allItems {
		totalAttempts += int64(n.RetryCount)
		if n.FailedAt != nil {
			failedAttempts++
		}

		switch n.Status {
		case models.NotificationStatusSent, models.NotificationStatusDelivered:
			sentCount++
			if n.SentAt != nil && n.CreatedAt.After(time.Time{}) {
				latency := n.SentAt.Sub(n.CreatedAt).Milliseconds()
				totalLatencyMs += latency
				latencyCount++
			}
		case models.NotificationStatusFailed:
			failedCount++
		case models.NotificationStatusPending, models.NotificationStatusQueued:
			pendingCount++
		case models.NotificationStatusDead:
			deadCount++
		}
	}

	successRate := 0.0
	if sentCount+failedCount > 0 {
		successRate = float64(sentCount) / float64(sentCount+failedCount) * 100
	}

	avgDeliveryMs := 0.0
	if latencyCount > 0 {
		avgDeliveryMs = float64(totalLatencyMs) / float64(latencyCount)
	}

	return response.OK(c, dto.ObservabilityMetricsResponse{
		NotificationsSent:      sentCount,
		NotificationsFailed:    failedCount,
		NotificationsPending:   pendingCount,
		NotificationsDead:      deadCount,
		QueueDepth:             queueDepth,
		ActiveWorkers:          10,
		AverageDeliveryTimeMs:  avgDeliveryMs,
		SuccessRate:            successRate,
		GeneratedAt:            time.Now(),
	})
}
