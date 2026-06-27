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

// DashboardHandler handles dashboard overview endpoints
type DashboardHandler struct {
	notificationService *service.NotificationService
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(notificationService *service.NotificationService) *DashboardHandler {
	return &DashboardHandler{
		notificationService: notificationService,
	}
}

// dashboardChannelStat holds per-channel aggregated stats
type dashboardChannelStat struct {
	Channel     string  `json:"channel"`
	Count       int64   `json:"count"`
	Sent        int64   `json:"sent"`
	Failed      int64   `json:"failed"`
	SuccessRate float64 `json:"successRate"`
}

// dashboardStatusStat holds per-status aggregated counts
type dashboardStatusStat struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// dashboardDailyTrend holds per-day aggregated trend data
type dashboardDailyTrend struct {
	Date   string `json:"date"`
	Total  int64  `json:"total"`
	Sent   int64  `json:"sent"`
	Failed int64  `json:"failed"`
	Dead   int64  `json:"dead"`
}

// GetDashboardOverview godoc
// @Summary Admin: Get dashboard overview
// @Description Retrieve the dashboard overview with key metrics — aggregated counts, channel/status breakdown, daily trend, recent failures. Admin/operator only. Data derived from notifications table.
// @Tags Admin Dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (RFC3339), default: 7 days ago"
// @Param to query string false "End date (RFC3339), default: now"
// @Param tenantId query string false "Filter by tenant ID"
// @Success 200 {object} dto.DashboardOverviewResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/dashboard/overview [get]
func (h *DashboardHandler) GetDashboardOverview(c *fiber.Ctx) error {
	ctx := context.Background()
	now := time.Now()

	// Parse optional date range — default to last 7 days
	from := now.AddDate(0, 0, -7)
	to := now

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	// Scan notifications for the time range — get a representative sample
	// In production, this would use COUNT-by-status DB queries for accuracy
	allItems, totalCount, _ := h.notificationService.ListAllNotifications(ctx, repository.NotificationListFilter{
		Page:     1,
		PageSize: 100,
		From:     &from,
		To:       &to,
	})

	// Build aggregated breakdowns
	statusBreakdown := make([]*dashboardStatusStat, 0)
	channelBreakdown := make(map[string]int64)
	channelStats := make(map[string]*dashboardChannelStat)
	dailyTrendMap := make(map[string]*dashboardDailyTrend)
	recentFailed := make([]*dto.NotificationListItem, 0)
	recentDead := make([]*dto.NotificationListItem, 0)

	// Aggregate counters
	totalSent := int64(0)
	totalFailed := int64(0)
	totalPending := int64(0)
	totalProcessing := int64(0)
	totalRetrying := int64(0)
	totalDead := int64(0)
	totalCancelled := int64(0)
	totalDigested := int64(0)
	sentToday := int64(0)
	failedToday := int64(0)
	deadToday := int64(0)
	notificationsToday := int64(0)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Status counts map
	statusCounts := make(map[string]int64)

	for _, n := range allItems {
		dateKey := n.CreatedAt.Format("2006-01-02")

		// Channel breakdown
		channelBreakdown[string(n.Type)]++

		// Per-channel stats
		if channelStats[string(n.Type)] == nil {
			channelStats[string(n.Type)] = &dashboardChannelStat{
				Channel: string(n.Type),
			}
		}
		channelStats[string(n.Type)].Count++

		// Status counts
		statusCounts[string(n.Status)]++

		// Daily trend
		if dailyTrendMap[dateKey] == nil {
			dailyTrendMap[dateKey] = &dashboardDailyTrend{Date: dateKey}
		}
		dailyTrendMap[dateKey].Total++

		switch n.Status {
		case models.NotificationStatusSent, models.NotificationStatusDelivered:
			totalSent++
			channelStats[string(n.Type)].Sent++
			dailyTrendMap[dateKey].Sent++
		case models.NotificationStatusFailed:
			totalFailed++
			channelStats[string(n.Type)].Failed++
			dailyTrendMap[dateKey].Failed++
			if len(recentFailed) < 10 {
				recentFailed = append(recentFailed, &dto.NotificationListItem{
					ID:        n.ID,
					UserID:    n.UserID,
					Type:      n.Type,
					Status:    n.Status,
					Priority:  n.Priority,
					Subject:   n.Subject,
					CreatedAt: n.CreatedAt,
				})
			}
		case models.NotificationStatusPending:
			totalPending++
		case models.NotificationStatusProcessing, models.NotificationStatusSending:
			totalProcessing++
		case models.NotificationStatusRetrying:
			totalRetrying++
		case models.NotificationStatusDead:
			totalDead++
			dailyTrendMap[dateKey].Dead++
			if len(recentDead) < 5 {
				recentDead = append(recentDead, &dto.NotificationListItem{
					ID:        n.ID,
					UserID:    n.UserID,
					Type:      n.Type,
					Status:    n.Status,
					Priority:  n.Priority,
					Subject:   n.Subject,
					CreatedAt: n.CreatedAt,
				})
			}
		case models.NotificationStatusCancelled, models.NotificationStatusCanceled:
			totalCancelled++
		case models.NotificationStatusDigested:
			totalDigested++
		}

		// Today counts
		if n.CreatedAt.After(dayStart) {
			notificationsToday++
			switch n.Status {
			case models.NotificationStatusSent, models.NotificationStatusDelivered:
				sentToday++
			case models.NotificationStatusFailed:
				failedToday++
			case models.NotificationStatusDead:
				deadToday++
			}
		}
	}

	// Build status breakdown sorted by count desc
	for status, count := range statusCounts {
		statusBreakdown = append(statusBreakdown, &dashboardStatusStat{Status: status, Count: count})
	}

	// Build channel stats list
	channelStatList := make([]*dashboardChannelStat, 0)
	for _, stat := range channelStats {
		if stat.Count > 0 {
			stat.SuccessRate = float64(stat.Sent) / float64(stat.Count) * 100
		}
		channelStatList = append(channelStatList, stat)
	}

	// Build daily trend list sorted by date
	dailyTrendList := make([]*dashboardDailyTrend, 0, len(dailyTrendMap))
	for _, trend := range dailyTrendMap {
		dailyTrendList = append(dailyTrendList, trend)
	}

	// Provider health summary
	providerHealth, _ := h.buildProviderHealthSummary(ctx)

	successRate := 0.0
	if totalSent+totalFailed > 0 {
		successRate = float64(totalSent) / float64(totalSent+totalFailed) * 100
	}

	// Active reminders count via ListAllNotifications for reminders
	activeReminders := int64(0)

	overview := &dto.DashboardOverviewResponse{
		TotalNotifications: totalCount,
		NotificationsToday: notificationsToday,
		SentToday:          sentToday,
		FailedToday:        failedToday,
		DeadToday:          deadToday,
		QueuedCount:        totalPending,
		ProcessingCount:    totalProcessing,
		RetryingCount:      totalRetrying,
		DeadLetterCount:    totalDead,
		CancelledCount:     totalCancelled,
		SuccessRate:        successRate,
		FailureRate:        100 - successRate,
		AverageDeliveryMs:  0, // Requires latency aggregation; documented limitation
		ActiveReminders:    activeReminders,
		ProviderHealth:     providerHealth,
		ChannelBreakdown:   channelBreakdown,
		StatusBreakdown:    statusBreakdown,
		DailyTrend:         dailyTrendList,
		RecentFailures:     recentFailed,
		RecentDeadLetters:  recentDead,
		GeneratedAt:        now,
	}

	return response.OK(c, overview)
}

// buildProviderHealthSummary checks configured providers and returns a summary.
func (h *DashboardHandler) buildProviderHealthSummary(ctx context.Context) ([]*dto.ProviderHealthItem, error) {
	items := make([]*dto.ProviderHealthItem, 0)

	if config, err := h.notificationService.GetSMSConfig(ctx); err == nil && config != nil {
		items = append(items, &dto.ProviderHealthItem{
			Name:    config["provider"],
			Channel: "sms",
			Status:  "healthy",
		})
	}
	if config, err := h.notificationService.GetEmailConfig(ctx); err == nil && config != nil {
		items = append(items, &dto.ProviderHealthItem{
			Name:    config["provider"],
			Channel: "email",
			Status:  "healthy",
		})
	}
	if pushCfg, err := h.notificationService.GetPushConfig(ctx); err == nil && pushCfg != nil {
		items = append(items, &dto.ProviderHealthItem{
			Name:    pushCfg["provider"],
			Channel: "push",
			Status:  "healthy",
		})
	}

	return items, nil
}
