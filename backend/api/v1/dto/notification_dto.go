package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/minisource/notifier/internal/models"
)

// == Standard Error Response ==

// ErrorResponse is the standard API error response
type ErrorResponse struct {
	Error     ErrorBody `json:"error"`
	RequestID string    `json:"requestId,omitempty"`
}

// ErrorBody contains the error details
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// == Standard Pagination ==

// PaginatedResponse is a generic paginated list response
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// NotificationRecipientRequest represents the recipient info for a notification
type NotificationRecipientRequest struct {
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
	UserID      string `json:"userId,omitempty"`
	DeviceToken string `json:"deviceToken,omitempty"`
	WebhookURL  string `json:"webhookUrl,omitempty"`
}

// CreateNotificationRequest represents a request to create a notification
// Canonical format uses channel + recipient. Legacy flat fields are also accepted.
type CreateNotificationRequest struct {
	// Channel is the primary notification channel (sms, email, push, in_app, webhook)
	Channel        models.NotificationChannel       `json:"channel,omitempty"`
	// Type is a backward-compatible alias for channel
	Type           models.NotificationType          `json:"type,omitempty"`
	// Priority defaults to normal
	Priority       models.NotificationPriority      `json:"priority,omitempty"`
	// Recipient is the canonical nested recipient object
	Recipient      *NotificationRecipientRequest    `json:"recipient,omitempty"`
	// Legacy flat recipient fields (backward compat)
	UserID         string                           `json:"userId,omitempty"`
	RecipientEmail string                           `json:"recipientEmail,omitempty"`
	RecipientPhone string                           `json:"recipientPhone,omitempty"`
	RecipientID    string                           `json:"recipientId,omitempty"`
	Subject        string                           `json:"subject,omitempty"`
	Body           string                           `json:"body,omitempty"`
	Metadata       map[string]interface{}           `json:"metadata,omitempty"`
	TemplateID     *uuid.UUID                       `json:"templateId,omitempty"`
	TemplateKey    string                           `json:"templateKey,omitempty"`
	Locale         string                           `json:"locale,omitempty"`
	ScheduledAt    *time.Time                       `json:"scheduledAt,omitempty"`
	IdempotencyKey string                           `json:"idempotencyKey,omitempty"`
	ProviderID     string                           `json:"providerId,omitempty"`
}

// Common error codes used across the API
const (
	ErrorCodeValidation     = "VALIDATION_ERROR"
	ErrorCodeUnauthorized   = "UNAUTHORIZED"
	ErrorCodeForbidden      = "FORBIDDEN"
	ErrorCodeNotFound       = "NOT_FOUND"
	ErrorCodeConflict       = "CONFLICT"
	ErrorCodeRateLimited    = "RATE_LIMITED"
	ErrorCodeProvider       = "PROVIDER_ERROR"
	ErrorCodeNotImplemented = "NOT_IMPLEMENTED"
	ErrorCodeInternal       = "INTERNAL_ERROR"
)

// == Delivery & Attempt DTOs ==

// DeliveryResponse represents a notification delivery attempt group
type DeliveryResponse struct {
	ID              uuid.UUID          `json:"id"`
	NotificationID  uuid.UUID          `json:"notificationId"`
	Provider        string             `json:"provider"`
	Channel         string             `json:"channel"`
	Status          string             `json:"status"`
	AttemptCount    int                `json:"attemptCount"`
	MaxAttempts     int                `json:"maxAttempts"`
	LastErrorCode   string             `json:"lastErrorCode,omitempty"`
	LastErrorMessage string            `json:"lastErrorMessage,omitempty"`
	NextRetryAt     *time.Time         `json:"nextRetryAt,omitempty"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
	CompletedAt     *time.Time         `json:"completedAt,omitempty"`
	Attempts        []*AttemptResponse `json:"attempts,omitempty"`
}

// AttemptResponse represents a single delivery attempt
type AttemptResponse struct {
	ID                     uuid.UUID  `json:"id"`
	DeliveryID             uuid.UUID  `json:"deliveryId"`
	AttemptNumber          int        `json:"attemptNumber"`
	Status                 string     `json:"status"`
	Retryable              bool       `json:"retryable"`
	ErrorCode              string     `json:"errorCode,omitempty"`
	ErrorMessage           string     `json:"errorMessage,omitempty"`
	ProviderMessageID      string     `json:"providerMessageId,omitempty"`
	ProviderResponseSanitized string `json:"providerResponseSanitized,omitempty"`
	LatencyMs              int64      `json:"latencyMs,omitempty"`
	CreatedAt              time.Time  `json:"createdAt"`
	CompletedAt            *time.Time `json:"completedAt,omitempty"`
	NextRetryAt            *time.Time `json:"nextRetryAt,omitempty"`
}

// == Reminder DTOs ==

// CreateReminderRequest represents a request to create a reminder
type CreateReminderRequest struct {
	UserID      uuid.UUID               `json:"userId" validate:"required"`
	Type        models.NotificationType `json:"type" validate:"required"`
	Recipient   string                  `json:"recipient" validate:"required"`
	TemplateID  *uuid.UUID              `json:"templateId"`
	TemplateKey string                  `json:"templateKey"`
	Locale      string                  `json:"locale"`
	Subject     string                  `json:"subject"`
	Body        string                  `json:"body"`
	Variables   map[string]string       `json:"variables"`
	ScheduledAt time.Time               `json:"scheduledAt" validate:"required"`
}

// ReminderResponse represents a reminder
type ReminderResponse struct {
	ID           uuid.UUID               `json:"id"`
	UserID       uuid.UUID               `json:"userId"`
	Type         models.NotificationType `json:"type"`
	Recipient    string                  `json:"recipient"`
	TemplateID   *uuid.UUID              `json:"templateId,omitempty"`
	TemplateKey  string                  `json:"templateKey,omitempty"`
	Locale       string                  `json:"locale,omitempty"`
	Subject      string                  `json:"subject,omitempty"`
	Body         string                  `json:"body,omitempty"`
	Variables    map[string]string       `json:"variables,omitempty"`
	ScheduledAt  time.Time               `json:"scheduledAt"`
	Status       string                  `json:"status"`
	NotifID      *uuid.UUID              `json:"notificationId,omitempty"`
	CreatedAt    time.Time               `json:"createdAt"`
	UpdatedAt    time.Time               `json:"updatedAt"`
	CancelledAt  *time.Time              `json:"cancelledAt,omitempty"`
}

// == Provider DTOs ==

// ProviderResponse represents a notification provider
type ProviderResponse struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Channel          string                 `json:"channel"`
	Type             string                 `json:"type,omitempty"`
	Status           string                 `json:"status"`
	IsEnabled        bool                   `json:"isEnabled"`
	IsPrimary        bool                   `json:"isPrimary"`
	IsDefault        bool                   `json:"isDefault"`
	Priority         int                    `json:"priority"`
	Description      string                 `json:"description,omitempty"`
	Config           map[string]interface{} `json:"config,omitempty"`
	SuccessRate      float64                `json:"successRate,omitempty"`
	AverageLatencyMs int64                  `json:"averageLatencyMs,omitempty"`
	LastSuccessAt    *time.Time             `json:"lastSuccessAt,omitempty"`
	LastFailureAt    *time.Time             `json:"lastFailureAt,omitempty"`
	LastError        string                 `json:"lastError,omitempty"`
	CreatedAt        time.Time              `json:"createdAt"`
	UpdatedAt        time.Time              `json:"updatedAt"`
}

// CreateProviderRequest represents a request to create a provider
type CreateProviderRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Channel     string                 `json:"channel" validate:"required"`
	Type        string                 `json:"type,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Priority    int                    `json:"priority,omitempty"`
	IsDefault   bool                   `json:"isDefault,omitempty"`
	Description string                 `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	SecretConfig map[string]interface{} `json:"secretConfig,omitempty"`
}

// UpdateProviderRequest represents a request to update a provider
type UpdateProviderRequest struct {
	Name         string                 `json:"name,omitempty"`
	Channel      string                 `json:"channel,omitempty"`
	Type         string                 `json:"type,omitempty"`
	Status       *string                `json:"status,omitempty"`
	Priority     *int                   `json:"priority,omitempty"`
	IsEnabled    *bool                  `json:"isEnabled,omitempty"`
	IsDefault    *bool                  `json:"isDefault,omitempty"`
	Description  *string                `json:"description,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	SecretConfig map[string]interface{} `json:"secretConfig,omitempty"`
}

// ToggleProviderStatusRequest represents a status toggle request
type ToggleProviderStatusRequest struct {
	IsEnabled bool   `json:"isEnabled,omitempty"`
	Status    string `json:"status,omitempty"`
}

// SetDefaultProviderRequest represents a request to set a provider as default
type SetDefaultProviderRequest struct {
	IsDefault bool `json:"isDefault"`
}

// ProviderListResponse represents a paginated provider list
type ProviderListResponse struct {
	Items      []*ProviderResponse `json:"items"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"pageSize"`
	TotalPages int                 `json:"totalPages"`
}

// == Dashboard DTOs ==

// DashboardOverviewResponse represents the dashboard overview
type DashboardOverviewResponse struct {
	TotalNotifications int64                         `json:"totalNotifications"`
	NotificationsToday int64                         `json:"notificationsToday"`
	SentToday          int64                         `json:"sentToday"`
	FailedToday        int64                         `json:"failedToday"`
	DeadToday          int64                         `json:"deadToday,omitempty"`
	QueuedCount        int64                         `json:"queuedCount"`
	ProcessingCount    int64                         `json:"processingCount,omitempty"`
	RetryingCount      int64                         `json:"retryingCount"`
	DeadLetterCount    int64                         `json:"deadLetterCount"`
	CancelledCount     int64                         `json:"cancelledCount,omitempty"`
	SuccessRate        float64                       `json:"successRate"`
	FailureRate        float64                       `json:"failureRate,omitempty"`
	AverageDeliveryMs  float64                       `json:"averageDeliveryMs"`
	ActiveReminders    int64                         `json:"activeReminders"`
	ProviderHealth     []*ProviderHealthItem         `json:"providerHealth,omitempty"`
	ChannelBreakdown   map[string]int64              `json:"channelBreakdown,omitempty"`
	StatusBreakdown    interface{}                   `json:"statusBreakdown,omitempty"`
	DailyTrend         interface{}                   `json:"dailyTrend,omitempty"`
	RecentNotifications []*NotificationListItem      `json:"recentNotifications,omitempty"`
	RecentFailures     []*NotificationListItem       `json:"recentFailures,omitempty"`
	RecentDeadLetters  []*NotificationListItem       `json:"recentDeadLetters,omitempty"`
	GeneratedAt        time.Time                     `json:"generatedAt"`
}

// ProviderHealthItem represents a single provider's health status
type ProviderHealthItem struct {
	Name    string `json:"name"`
	Channel string `json:"channel"`
	Status  string `json:"status"`
	SuccessRate float64 `json:"successRate"`
}

// ProviderHealthResponse represents the aggregate health of all providers
type ProviderHealthResponse struct {
	Providers     []*ProviderHealthItem `json:"providers"`
	HealthyCount  int64                 `json:"healthyCount"`
	DegradedCount int64                 `json:"degradedCount"`
	DownCount     int64                 `json:"downCount"`
	DisabledCount int64                 `json:"disabledCount"`
	CheckedAt     time.Time             `json:"checkedAt"`
}

// ProviderTestRequest represents a provider test request
type ProviderTestRequest struct {
	Channel   string `json:"channel,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Subject   string `json:"subject,omitempty"`
	Body      string `json:"body,omitempty"`
	DryRun    bool   `json:"dryRun"`
}

// ProviderTestResponse represents a provider test result
type ProviderTestResponse struct {
	ProviderID                 string    `json:"providerId"`
	Channel                    string    `json:"channel,omitempty"`
	DryRun                     bool      `json:"dryRun"`
	Success                    bool      `json:"success"`
	Status                     string    `json:"status"`
	Message                    string    `json:"message,omitempty"`
	ProviderMessageID          string    `json:"providerMessageId,omitempty"`
	ProviderResponseSanitized  string    `json:"providerResponseSanitized,omitempty"`
	LatencyMs                  int64     `json:"latencyMs,omitempty"`
	CheckedAt                  time.Time `json:"checkedAt"`
}

// == Observability DTOs ==

// DependencyHealth represents the health of a single dependency
type DependencyHealth struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// ObservabilityHealthResponse represents the health check response
type ObservabilityHealthResponse struct {
	Status        string              `json:"status"`
	Service       string              `json:"service,omitempty"`
	Version       string              `json:"version"`
	Environment   string              `json:"environment,omitempty"`
	Uptime        string              `json:"uptime,omitempty"`
	UptimeSeconds int64               `json:"uptimeSeconds,omitempty"`
	Dependencies  []*DependencyHealth `json:"dependencies,omitempty"`
	Timestamp     time.Time           `json:"timestamp"`
	GeneratedAt   time.Time           `json:"generatedAt,omitempty"`
}

// ReadinessCheck represents the result of a single readiness check
type ReadinessCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// ObservabilityReadinessResponse represents the readiness check response
type ObservabilityReadinessResponse struct {
	Ready       bool               `json:"ready"`
	Overall     string             `json:"status,omitempty"`
	Status      string             `json:"overall,omitempty"`
	Checks      []*ReadinessCheck  `json:"checks"`
	Timestamp   time.Time          `json:"timestamp"`
	GeneratedAt time.Time          `json:"generatedAt,omitempty"`
}

// ObservabilityMetricsResponse represents operational metrics
type ObservabilityMetricsResponse struct {
	NotificationsSent      int64   `json:"notificationsSent"`
	NotificationsFailed    int64   `json:"notificationsFailed"`
	NotificationsPending   int64   `json:"notificationsPending"`
	NotificationsDead      int64   `json:"notificationsDead"`
	QueueDepth             int64   `json:"queueDepth"`
	ActiveWorkers          int     `json:"activeWorkers"`
	AverageDeliveryTimeMs  float64 `json:"averageDeliveryTimeMs"`
	SuccessRate            float64 `json:"successRate"`
	TotalAttempts          int64   `json:"totalAttempts,omitempty"`
	FailedAttempts         int64   `json:"failedAttempts,omitempty"`
	GeneratedAt            time.Time `json:"generatedAt"`
}

// == Queue / Worker Observability DTOs ==

// QueueOverviewResponse represents the state of the notification queue
type QueueOverviewResponse struct {
	PendingCount      int64      `json:"pendingCount"`
	ProcessingCount   int64      `json:"processingCount"`
	RetryingCount     int64      `json:"retryingCount"`
	DeadCount         int64      `json:"deadCount"`
	ScheduledCount    int64      `json:"scheduledCount"`
	OldestPendingAt   *time.Time `json:"oldestPendingAt,omitempty"`
	NextRetryAt       *time.Time `json:"nextRetryAt,omitempty"`
	ThroughputPerMin  float64    `json:"throughputPerMinute"`
	AverageLatencyMs  float64    `json:"averageLatencyMs"`
	GeneratedAt       time.Time  `json:"generatedAt"`
}

// WorkerInfo represents a worker's status
type WorkerInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Channel   string    `json:"channel,omitempty"`
	QueueSize int       `json:"queueSize"`
}

// WorkerOverviewResponse represents the state of all workers
type WorkerOverviewResponse struct {
	Workers          []*WorkerInfo `json:"workers"`
	ActiveCount      int           `json:"activeCount"`
	IdleCount        int           `json:"idleCount"`
	FailedCount      int           `json:"failedCount,omitempty"`
	LastHeartbeatAt  *time.Time    `json:"lastHeartbeatAt,omitempty"`
	GeneratedAt      time.Time     `json:"generatedAt"`
}

// BatchNotificationRequest represents a batch notification request
type BatchNotificationRequest struct {
	Notifications []CreateNotificationRequest `json:"notifications" validate:"required,min=1,max=100"`
}

// NotificationResponse represents a notification response
type NotificationResponse struct {
	ID             uuid.UUID                   `json:"id"`
	TenantID       *uuid.UUID                  `json:"tenantId,omitempty"`
	UserID         uuid.UUID                   `json:"userId"`
	Type           models.NotificationType     `json:"type"`
	Status         models.NotificationStatus   `json:"status"`
	Priority       models.NotificationPriority `json:"priority"`
	RecipientEmail string                      `json:"recipientEmail,omitempty"`
	RecipientPhone string                      `json:"recipientPhone,omitempty"`
	Subject        string                      `json:"subject,omitempty"`
	Body           string                      `json:"body"`
	TemplateID     *uuid.UUID                  `json:"templateId,omitempty"`
	TemplateKey    string                      `json:"templateKey,omitempty"`
	Locale         string                      `json:"locale,omitempty"`
	RetryCount     int                         `json:"retryCount"`
	MaxRetries     int                         `json:"maxRetries"`
	ErrorMessage   string                      `json:"errorMessage,omitempty"`
	Provider       string                      `json:"provider,omitempty"`
	ScheduledAt    *time.Time                  `json:"scheduledAt,omitempty"`
	SentAt         *time.Time                  `json:"sentAt,omitempty"`
	DeliveredAt    *time.Time                  `json:"deliveredAt,omitempty"`
	FailedAt       *time.Time                  `json:"failedAt,omitempty"`
	SeenAt         *time.Time                  `json:"seenAt,omitempty"`
	ReadAt         *time.Time                  `json:"readAt,omitempty"`
	ClickedAt      *time.Time                  `json:"clickedAt,omitempty"`
	CancelledAt    *time.Time                  `json:"cancelledAt,omitempty"`
	CreatedAt      time.Time                   `json:"createdAt"`
	UpdatedAt      time.Time                   `json:"updatedAt"`
}

// NotificationListItem is a compact notification response for list views
type NotificationListItem struct {
	ID             uuid.UUID                   `json:"id"`
	TenantID       *uuid.UUID                  `json:"tenantId,omitempty"`
	UserID         uuid.UUID                   `json:"userId"`
	Type           models.NotificationType     `json:"type"`
	Status         models.NotificationStatus   `json:"status"`
	Priority       models.NotificationPriority `json:"priority"`
	RecipientEmail string                      `json:"recipientEmail,omitempty"`
	RecipientPhone string                      `json:"recipientPhone,omitempty"`
	Subject        string                      `json:"subject,omitempty"`
	BodyPreview    string                      `json:"bodyPreview,omitempty"`
	TemplateKey    string                      `json:"templateKey,omitempty"`
	Locale         string                      `json:"locale,omitempty"`
	RetryCount     int                         `json:"retryCount"`
	CreatedAt      time.Time                   `json:"createdAt"`
	UpdatedAt      time.Time                   `json:"updatedAt"`
	SentAt         *time.Time                  `json:"sentAt,omitempty"`
}

// PaginatedNotificationResponse represents paginated notifications (backward compatible with existing handlers)
type PaginatedNotificationResponse struct {
	Data       []*NotificationResponse `json:"data"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"pageSize"`
	TotalPages int                     `json:"totalPages"`
}

// UpdatePreferenceRequest represents a notification preference update
type UpdatePreferenceRequest struct {
	Type             models.NotificationType          `json:"type" validate:"required"`
	IsEnabled        bool                             `json:"isEnabled"`
	AllowInstant     bool                             `json:"allowInstant"`
	AllowDigest      bool                             `json:"allowDigest"`
	DigestFrequency  string                           `json:"digestFrequency"`
	QuietHours       *models.QuietHoursConfig         `json:"quietHours,omitempty"`
	CategorySettings map[models.NotificationCategory]bool `json:"categorySettings,omitempty"`
}

// ChannelPreferenceRequest represents a per-channel preference update
type ChannelPreferenceRequest struct {
	IsEnabled       bool                       `json:"isEnabled"`
	AllowInstant    *bool                      `json:"allowInstant,omitempty"`
	AllowDigest     *bool                      `json:"allowDigest,omitempty"`
	DigestFrequency string                     `json:"digestFrequency,omitempty"`
	QuietHours      *models.QuietHoursConfig   `json:"quietHours,omitempty"`
}

// PreferenceResponse represents a notification preference response
type PreferenceResponse struct {
	ID               uuid.UUID                       `json:"id"`
	UserID           uuid.UUID                       `json:"userId"`
	Type             models.NotificationType         `json:"type"`
	IsEnabled        bool                            `json:"isEnabled"`
	AllowInstant     bool                            `json:"allowInstant"`
	AllowDigest      bool                            `json:"allowDigest"`
	DigestFrequency  string                          `json:"digestFrequency"`
	QuietHours       *models.QuietHoursConfig        `json:"quietHours,omitempty"`
	CategorySettings map[models.NotificationCategory]bool `json:"categorySettings,omitempty"`
}

// UnreadCountResponse represents the unread notification count for a user
type UnreadCountResponse struct {
	UserID uuid.UUID `json:"userId"`
	Count  int64     `json:"unreadCount"`
}

// ActionResponse represents a generic action response
type ActionResponse struct {
	Message string      `json:"message"`
	ID      uuid.UUID   `json:"id,omitempty"`
	Status  string      `json:"status,omitempty"`
}

// MarkAllAsReadResponse represents the response from marking all notifications as read
type MarkAllAsReadResponse struct {
	Message      string    `json:"message"`
	UserID       uuid.UUID `json:"userId"`
	UpdatedCount int64     `json:"updatedCount"`
}

// CreateTemplateRequest represents a template creation request
type CreateTemplateRequest struct {
	Key              string                  `json:"key"`
	Name             string                  `json:"name" validate:"required"`
	Type             models.NotificationType `json:"type" validate:"required"`
	Locale           string                  `json:"locale"`
	Subject          string                  `json:"subject"`
	Body             string                  `json:"body" validate:"required"`
	Description      string                  `json:"description"`
	Variables        []string                `json:"variables"`
	Provider         string                  `json:"provider"`
	ProviderTemplate string                  `json:"providerTemplate"`
	IsActive         bool                    `json:"isActive"`
}

// TemplateResponse represents a template response
type TemplateResponse struct {
	ID          uuid.UUID               `json:"id"`
	Key         string                  `json:"key,omitempty"`
	Name        string                  `json:"name"`
	Type        models.NotificationType `json:"type"`
	Locale      string                  `json:"locale"`
	Subject     string                  `json:"subject,omitempty"`
	Body        string                  `json:"body"`
	Description string                  `json:"description,omitempty"`
	IsActive    bool                    `json:"isActive"`
	CreatedAt   time.Time               `json:"createdAt"`
}

// RenderPreviewRequest represents a request to render a template preview
type RenderPreviewRequest struct {
	TemplateKey string            `json:"templateKey,omitempty"` // For key-based render
	TemplateID  string            `json:"templateId,omitempty"` // For ID-based render (alternative)
	Variables   map[string]string `json:"variables"`
	Locale      string            `json:"locale"`
}

// RenderPreviewResponse represents a rendered template preview result
type RenderPreviewResponse struct {
	Subject         string   `json:"subject,omitempty"`
	Body            string   `json:"body"`
	UsedVariables   []string `json:"usedVariables,omitempty"`
	MissingVariables []string `json:"missingVariables,omitempty"`
}
