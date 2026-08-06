package worker

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
)

// NotificationJob represents a notification job to be processed
type NotificationJob struct {
	Notification *models.Notification
	Retries      int
}

// NotificationWorker handles asynchronous notification processing
type NotificationWorker struct {
	jobQueue      chan *NotificationJob
	workers       int
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	logger        logging.Logger
	config        *config.Config
	notifRepo     repository.NotificationRepository
	logRepo       repository.NotificationLogRepository
	attemptRepo   repository.ProviderAttemptRepository
	smsHandler    SMSHandler
	emailHandler  EmailHandler
	pushHandler   PushHandler
	digestService DigestProcessor
	deliveryControl DeliveryControl}

// DigestProcessor interface for processing digest notifications
type DigestProcessor interface {
	ProcessDueDigests(ctx context.Context)
}

// DeliveryControl is the global outbound-delivery pause reader used by the
// worker. Defined here (not in the service package) to avoid an import cycle:
// the service package imports the worker, so the worker must only depend on
// this interface. The service's DeliveryControlService satisfies it.
type DeliveryControl interface {
	// IsPaused reports whether outbound provider execution must stop.
	IsPaused(ctx context.Context) bool
	// HoldNotification freezes one delivery (status=held) without consuming
	// its retry budget.
	HoldNotification(ctx context.Context, notificationID uuid.UUID, reason string) error
	// ReleaseHeld re-queues up to limit held deliveries (controlled release).
	ReleaseHeld(ctx context.Context, limit int) (int64, error)
	// CheckAutoResume resumes automatically when a pause deadline expired.
	CheckAutoResume(ctx context.Context) error
	// PurgeExpiredIdempotency removes expired pause/resume idempotency records.
	PurgeExpiredIdempotency(ctx context.Context) (int64, error)
}


// SMSHandler interface for sending SMS
type SMSHandler interface {
	SendSMS(ctx context.Context, notification *models.Notification) (string, error)
}

// EmailHandler interface for sending emails
type EmailHandler interface {
	SendEmail(ctx context.Context, notification *models.Notification) (string, error)
}

// PushHandler interface for sending push notifications
type PushHandler interface {
	SendPush(ctx context.Context, notification *models.Notification) (string, error)
}

// NewNotificationWorker creates a new notification worker
func NewNotificationWorker(
	cfg *config.Config,
	logger logging.Logger,
	notifRepo repository.NotificationRepository,
	logRepo repository.NotificationLogRepository,
	attemptRepo repository.ProviderAttemptRepository,
	smsHandler SMSHandler,
	emailHandler EmailHandler,
	pushHandler PushHandler,
	digestProcessor DigestProcessor,
	deliveryControl DeliveryControl,
) *NotificationWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &NotificationWorker{
		jobQueue:      make(chan *NotificationJob, cfg.Worker.QueueSize),
		workers:       cfg.Worker.NumWorkers,
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger,
		config:        cfg,
		notifRepo:     notifRepo,
		logRepo:       logRepo,
		attemptRepo:   attemptRepo,
		smsHandler:    smsHandler,
		emailHandler:  emailHandler,
		pushHandler:   pushHandler,
		digestService: digestProcessor,
		deliveryControl: deliveryControl,
	}
}

// Start starts the worker pool
func (w *NotificationWorker) Start() {
	w.logger.Info(logging.General, logging.Startup, "Starting notification workers", map[logging.ExtraKey]interface{}{
		"numWorkers":    w.workers,
		"queueSize":     cap(w.jobQueue),
		"pollEnabled":   w.config.Worker.PollEnabled,
		"pollInterval":  w.config.Worker.PollInterval,
	})

	// Start worker goroutines
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.worker(i)
	}

	// Start periodic retry processor
	w.wg.Add(1)
	go w.retryProcessor()

	// Start pending notification poller (DB-backed queue recovery)
	if w.config.Worker.PollEnabled {
		w.wg.Add(1)
		go w.pendingPoller()
		w.wg.Add(1)
		go w.queueDepthLogger()
	}

	// Start digest processor (accumulated batch delivery)
	if w.config.Digest.Enabled {
		w.wg.Add(1)
		go w.digestProcessor()
	}

	// Start provider attempt retention cleanup (provider request lifecycle logs)
	if w.attemptRepo != nil {
		w.wg.Add(1)
		go w.attemptCleanupProcessor()
	}

	// Start global delivery-pause control processor: auto-resume on deadline
	// and controlled release of held work (no thundering herd).
	if w.deliveryControl != nil {
		w.wg.Add(1)
		go w.deliveryControlProcessor()
	}

	w.logger.Info(logging.General, logging.Startup, "Notification workers started successfully", nil)
}

// Stop stops the worker pool gracefully
func (w *NotificationWorker) Stop() {
	w.logger.Info(logging.General, logging.Startup, "Stopping notification workers", nil)

	w.cancel()
	close(w.jobQueue)
	w.wg.Wait()

	w.logger.Info(logging.General, logging.Startup, "Notification workers stopped successfully", nil)
}

// EnqueueNotification adds a notification to the processing queue
func (w *NotificationWorker) EnqueueNotification(notification *models.Notification) error {
	w.logger.Debug(logging.General, logging.Insert, "Enqueueing notification", map[logging.ExtraKey]interface{}{
		"notificationId": notification.ID,
		"type":           notification.Type,
		"priority":       notification.Priority,
	})

	select {
	case w.jobQueue <- &NotificationJob{Notification: notification, Retries: 0}:
		w.logger.Debug(logging.General, logging.Insert, "Notification enqueued successfully", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
		})
		return nil
	case <-w.ctx.Done():
		w.logger.Warn(logging.General, logging.Insert, "Worker is shutting down, cannot enqueue", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
		})
		return w.ctx.Err()
	default:
		w.logger.Warn(logging.General, logging.Insert, "Job queue is full", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"queueSize":      cap(w.jobQueue),
		})
		return ErrQueueFull
	}
}

// SendNotificationSync sends a notification synchronously without queueing
// This bypasses the worker queue and sends immediately, returning real errors
func (w *NotificationWorker) SendNotificationSync(ctx context.Context, notification *models.Notification) error {
	w.logger.Debug(logging.Internal, logging.Api, "Sending notification synchronously", map[logging.ExtraKey]interface{}{
		"notificationId": notification.ID,
		"type":           notification.Type,
	})

	// Pause gate — synchronous sends are also frozen.
	if w.deliveryControl != nil && w.deliveryControl.IsPaused(ctx) {
		w.holdForPause(ctx, notification, "sync_send")
		return ErrDeliveryPaused
	}

	var providerMsgID string
	var err error

	switch notification.Type {
	case models.NotificationTypeSMS:
		providerMsgID, err = w.smsHandler.SendSMS(ctx, notification)
	case models.NotificationTypeEmail:
		providerMsgID, err = w.emailHandler.SendEmail(ctx, notification)
	case models.NotificationTypePush, models.NotificationTypeInApp:
		providerMsgID, err = w.pushHandler.SendPush(ctx, notification)
	default:
		return ErrUnsupportedNotificationType
	}

	if err != nil {
		w.logger.Error(logging.Internal, logging.Api, "Failed to send notification (sync)", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"type":           notification.Type,
			"error":          err.Error(),
		})
		return err
	}

	w.logger.Info(logging.Internal, logging.Api, "Notification sent successfully (sync)", map[logging.ExtraKey]interface{}{
		"notificationId": notification.ID,
		"type":           notification.Type,
		"providerMsgId":  providerMsgID,
	})

	return nil
}

// deliveryControlProcessor periodically checks for auto-resume and releases a
// bounded batch of held deliveries so a large backlog resumes gradually.
func (w *NotificationWorker) deliveryControlProcessor() {
	defer w.wg.Done()

	interval := 5 * time.Second
	if w.config.DeliveryControl.ReleaseIntervalSec > 0 {
		interval = time.Duration(w.config.DeliveryControl.ReleaseIntervalSec) * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	w.logger.Debug(logging.General, logging.Startup, "Delivery control processor started", map[logging.ExtraKey]interface{}{
		"interval": interval.String(),
	})

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			// 1) Auto-resume when an optional pause deadline passed.
			if err := w.deliveryControl.CheckAutoResume(ctx); err != nil {
				w.logger.Warn(logging.General, logging.Update, "Delivery control auto-resume check failed", map[logging.ExtraKey]interface{}{
					"error": err.Error(),
				})
			}
			// 2) Controlled release of held deliveries (no-op while paused).
			if _, err := w.deliveryControl.ReleaseHeld(ctx, w.config.DeliveryControl.ReleaseBatchSize); err != nil {
				w.logger.Warn(logging.General, logging.Update, "Delivery control release failed", map[logging.ExtraKey]interface{}{
					"error": err.Error(),
				})
			}
			// 3) Lease recovery: re-queue deliveries stuck in sending/processing
			// past the age threshold after a worker crash. Never consumes retry
			// budget; the DB poller picks them up again.
			if age := w.config.DeliveryControl.StuckSendingAgeSec; age > 0 {
				cutoff := time.Now().Add(-time.Duration(age) * time.Second)
				if recovered, rerr := w.notifRepo.RecoverStuckSending(ctx, cutoff); rerr != nil {
					w.logger.Warn(logging.General, logging.Update, "Stuck sending recovery failed", map[logging.ExtraKey]interface{}{
						"error": rerr.Error(),
					})
				} else if recovered > 0 {
					w.logger.Info(logging.General, logging.Update, "Recovered stuck sending deliveries", map[logging.ExtraKey]interface{}{
						"count": recovered,
					})
				}
			}
			// 4) Bounded retention: purge expired pause/resume idempotency rows.
			if _, perr := w.deliveryControl.PurgeExpiredIdempotency(ctx); perr != nil {
				w.logger.Warn(logging.General, logging.Delete, "Idempotency purge failed", map[logging.ExtraKey]interface{}{
					"error": perr.Error(),
				})
			}
		case <-w.ctx.Done():
			w.logger.Debug(logging.General, logging.Startup, "Delivery control processor shutting down", nil)
			return
		}
	}
}

// worker processes jobs from the queue
func (w *NotificationWorker) worker(id int) {
	defer w.wg.Done()

	w.logger.Debug(logging.General, logging.Startup, "Worker started", map[logging.ExtraKey]interface{}{
		"workerId": id,
	})

	for {
		select {
		case job, ok := <-w.jobQueue:
			if !ok {
				w.logger.Debug(logging.General, logging.Startup, "Worker shutting down", map[logging.ExtraKey]interface{}{
					"workerId": id,
				})
				return
			}

			w.processJob(id, job)

		case <-w.ctx.Done():
			w.logger.Debug(logging.General, logging.Startup, "Worker context cancelled", map[logging.ExtraKey]interface{}{
				"workerId": id,
			})
			return
		}
	}
}

// processJob processes a single notification job
func (w *NotificationWorker) processJob(workerID int, job *NotificationJob) {
	ctx := context.Background()
	startTime := time.Now()

	notification := job.Notification

	w.logger.Debug(logging.Internal, logging.Api, "Processing notification", map[logging.ExtraKey]interface{}{
		"workerId":       workerID,
		"notificationId": notification.ID,
		"type":           notification.Type,
		"retries":        job.Retries,
	})

	// FINAL authoritative pause gate — held before any provider work.
	if w.deliveryControl != nil && w.deliveryControl.IsPaused(ctx) {
		w.holdForPause(ctx, notification, "initial_send")
		return
	}

	// Update status to sending
	if err := w.notifRepo.UpdateStatus(ctx, notification.ID, models.NotificationStatusSending); err != nil {
		w.logger.Error(logging.General, logging.Update, "Failed to update notification status", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"error":          err.Error(),
		})
	}

	// Create log entry for sending
	w.createLog(ctx, notification.ID, "sending", models.NotificationStatusSending, "Attempting to send notification", "")

	// Send notification based on type
	var providerMsgID string
	var err error

	switch notification.Type {
	case models.NotificationTypeSMS:
		providerMsgID, err = w.smsHandler.SendSMS(ctx, notification)
	case models.NotificationTypeEmail:
		providerMsgID, err = w.emailHandler.SendEmail(ctx, notification)
	case models.NotificationTypePush, models.NotificationTypeInApp:
		providerMsgID, err = w.pushHandler.SendPush(ctx, notification)
	default:
		err = ErrUnsupportedNotificationType
	}

	processingTime := int(time.Since(startTime).Milliseconds())

	if err != nil {
		// Pause control errors are NOT provider failures: hold + preserve retry.
		if errors.Is(err, ErrDeliveryPaused) {
			w.holdForPause(ctx, notification, "provider_boundary_gate")
			return
		}
		w.handleFailure(ctx, notification, job, err, processingTime)
	} else {
		w.handleSuccess(ctx, notification, providerMsgID, processingTime)
	}
}

// handleSuccess handles successful notification sending
func (w *NotificationWorker) handleSuccess(ctx context.Context, notification *models.Notification, providerMsgID string, processingTime int) {
	w.logger.Info(logging.Internal, logging.Api, "Notification sent successfully", map[logging.ExtraKey]interface{}{
		"notificationId": notification.ID,
		"type":           notification.Type,
		"providerMsgId":  providerMsgID,
		"processingTime": processingTime,
	})

	// Mark as sent
	if err := w.notifRepo.MarkAsSent(ctx, notification.ID, providerMsgID); err != nil {
		w.logger.Error(logging.General, logging.Update, "Failed to mark notification as sent", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"error":          err.Error(),
		})
	}

	// Persist which provider handled this notification (set by the adapter).
	if err := w.notifRepo.SetProvider(ctx, notification.ID, notification.Provider); err != nil {
		w.logger.Debug(logging.General, logging.Update, "Failed to persist notification provider", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"provider":       notification.Provider,
			"error":          err.Error(),
		})
	}

	// Create success log
	w.createLog(ctx, notification.ID, "sent", models.NotificationStatusSent, "Notification sent successfully", "")
}

// holdForPause marks a delivery as held by the global pause WITHOUT consuming
// its retry budget, and records a "held" log entry.
func (w *NotificationWorker) holdForPause(ctx context.Context, notification *models.Notification, reason string) {
	w.logger.Info(logging.Internal, logging.Api, "Holding notification for global delivery pause", map[logging.ExtraKey]interface{}{
		"notificationId": notification.ID,
		"reason":        reason,
	})
	if err := w.deliveryControl.HoldNotification(ctx, notification.ID, reason); err != nil {
		w.logger.Error(logging.General, logging.Update, "Failed to hold notification for pause", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"error":          err.Error(),
		})
	}
	w.createLog(ctx, notification.ID, "held", models.NotificationStatusHeld, "Held by global outbound delivery pause", "")
}

// handleFailure handles failed notification sending with retry logic
func (w *NotificationWorker) handleFailure(ctx context.Context, notification *models.Notification, job *NotificationJob, err error, processingTime int) {
	_ = job            // job parameter reserved for future use
	_ = processingTime // processingTime parameter reserved for future metrics

	w.logger.Error(logging.Internal, logging.Api, "Failed to send notification", map[logging.ExtraKey]interface{}{
		"notificationId": notification.ID,
		"type":           notification.Type,
		"error":          err.Error(),
		"retryCount":     notification.RetryCount,
		"maxRetries":     notification.MaxRetries,
	})

	// Persist which provider failed this notification (set by the adapter) so
	// per-provider stats include failures, not just successes.
	if err := w.notifRepo.SetProvider(ctx, notification.ID, notification.Provider); err != nil {
		w.logger.Debug(logging.General, logging.Update, "Failed to persist notification provider (failure)", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"provider":       notification.Provider,
			"error":          err.Error(),
		})
	}

	// Check if we should retry
	if notification.RetryCount < notification.MaxRetries {
		// Calculate next retry time using exponential backoff
		nextRetryAt := w.calculateNextRetryTime(notification.RetryCount)

		// Update notification for retry
		if err := w.notifRepo.IncrementRetryCount(ctx, notification.ID, nextRetryAt, err.Error()); err != nil {
			w.logger.Error(logging.General, logging.Update, "Failed to update retry count", map[logging.ExtraKey]interface{}{
				"notificationId": notification.ID,
				"error":          err.Error(),
			})
		}

		w.logger.Info(logging.Internal, logging.Api, "Notification scheduled for retry", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"retryCount":     notification.RetryCount + 1,
			"nextRetryAt":    nextRetryAt,
		})

		// Create retry log
		w.createLog(ctx, notification.ID, "retrying", models.NotificationStatusRetrying, "Scheduled for retry", err.Error())
	} else {
		// Max retries reached — mark as dead-letter (canceled)
		if err := w.notifRepo.MarkAsDeadLetter(ctx, notification.ID, err.Error()); err != nil {
			w.logger.Error(logging.General, logging.Update, "Failed to mark notification as dead-letter", map[logging.ExtraKey]interface{}{
				"notificationId": notification.ID,
				"error":          err.Error(),
			})
		}

		w.logger.Warn(logging.Internal, logging.Api, "Notification moved to dead-letter after max retries", map[logging.ExtraKey]interface{}{
			"notificationId": notification.ID,
			"retryCount":     notification.RetryCount,
			"maxRetries":     notification.MaxRetries,
		})

		// Create dead-letter log entry
		w.createLog(ctx, notification.ID, "dead_letter", models.NotificationStatusCanceled, "Max retries exceeded, moved to dead-letter", err.Error())
	}
}

// calculateNextRetryTime calculates the next retry time using exponential backoff
func (w *NotificationWorker) calculateNextRetryTime(retryCount int) time.Time {
	baseDelay := time.Duration(w.config.Worker.RetryBaseDelay) * time.Second
	maxDelay := time.Duration(w.config.Worker.RetryMaxDelay) * time.Second

	// Exponential backoff: baseDelay * 2^retryCount
	delay := baseDelay * time.Duration(math.Pow(2, float64(retryCount)))

	// Cap at max delay
	if delay > maxDelay {
		delay = maxDelay
	}

	nextRetry := time.Now().Add(delay)

	w.logger.Debug(logging.Internal, logging.Api, "Calculated next retry time", map[logging.ExtraKey]interface{}{
		"retryCount":  retryCount,
		"delay":       delay.String(),
		"nextRetryAt": nextRetry,
	})

	return nextRetry
}

// retryProcessor periodically checks for notifications that need to be retried
func (w *NotificationWorker) retryProcessor() {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	w.logger.Debug(logging.General, logging.Startup, "Retry processor started", nil)

	for {
		select {
		case <-ticker.C:
			w.processRetries()
		case <-w.ctx.Done():
			w.logger.Debug(logging.General, logging.Startup, "Retry processor shutting down", nil)
			return
		}
	}
}

// processRetries fetches and re-enqueues notifications that are ready for retry
func (w *NotificationWorker) processRetries() {
	ctx := context.Background()

	w.logger.Debug(logging.Internal, logging.Api, "Processing retries", nil)

	// Retries are frozen while the global pause is active: hold them instead
	// of executing, preserving their retry budget.
	if w.deliveryControl != nil && w.deliveryControl.IsPaused(ctx) {
		notifications, err := w.notifRepo.GetRetryableNotifications(ctx, 100)
		if err == nil {
			for _, n := range notifications {
				w.holdForPause(ctx, n, "retry")
			}
		}
		return
	}

	// Fetch retryable notifications
	notifications, err := w.notifRepo.GetRetryableNotifications(ctx, 100)
	if err != nil {
		w.logger.Error(logging.General, logging.Select, "Failed to fetch retryable notifications", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return
	}

	if len(notifications) > 0 {
		w.logger.Info(logging.Internal, logging.Api, "Found notifications to retry", map[logging.ExtraKey]interface{}{
			"count": len(notifications),
		})

		for _, notification := range notifications {
			// Update status to sending before re-enqueueing
			notification.Status = models.NotificationStatusSending

			if err := w.EnqueueNotification(notification); err != nil {
				w.logger.Error(logging.General, logging.Insert, "Failed to enqueue retry", map[logging.ExtraKey]interface{}{
					"notificationId": notification.ID,
					"error":          err.Error(),
				})
			}
		}
	}
}

// pendingPoller periodically polls the database for pending notifications and enqueues them.
// This ensures queued notifications are not lost on server restart (DB-backed queue recovery).
func (w *NotificationWorker) pendingPoller() {
	defer w.wg.Done()

	interval := time.Duration(w.config.Worker.PollInterval) * time.Second
	if interval < 1*time.Second {
		interval = 15 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	w.logger.Debug(logging.General, logging.Startup, "Pending notification poller started", map[logging.ExtraKey]interface{}{
		"interval": interval.String(),
	})

	for {
		select {
		case <-ticker.C:
			w.processPending()
		case <-w.ctx.Done():
			w.logger.Debug(logging.General, logging.Startup, "Pending poller shutting down", nil)
			return
		}
	}
}

// processPending fetches pending notifications from DB and enqueues them
func (w *NotificationWorker) processPending() {
	ctx := context.Background()

	// New/scheduled deliveries are held while the global pause is active —
	// accepted, persisted, but never reaching a provider.
	if w.deliveryControl != nil && w.deliveryControl.IsPaused(ctx) {
		notifications, err := w.notifRepo.GetPendingNotifications(ctx, w.config.Worker.QueueSize)
		if err == nil {
			for _, n := range notifications {
				w.holdForPause(ctx, n, "scheduled_or_initial")
			}
		}
		return
	}

	notifications, err := w.notifRepo.GetPendingNotifications(ctx, w.config.Worker.QueueSize)
	if err != nil {
		w.logger.Error(logging.General, logging.Select, "Failed to fetch pending notifications", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return
	}

	if len(notifications) > 0 {
		w.logger.Info(logging.Internal, logging.Api, "Found pending notifications from DB", map[logging.ExtraKey]interface{}{
			"count": len(notifications),
		})

		for _, notification := range notifications {
			select {
			case w.jobQueue <- &NotificationJob{Notification: notification, Retries: 0}:
				// Successfully enqueued
			case <-w.ctx.Done():
				return
			default:
				w.logger.Warn(logging.General, logging.Insert, "Queue full, stopping pending poll cycle", nil)
				return
			}
		}
	}
}

// queueDepthLogger periodically logs the queue depth for monitoring
func (w *NotificationWorker) queueDepthLogger() {
	defer w.wg.Done()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			depth, err := w.notifRepo.GetQueueDepth(ctx)
			if err != nil {
				w.logger.Debug(logging.General, logging.Select, "Failed to get queue depth", map[logging.ExtraKey]interface{}{
					"error": err.Error(),
				})
				continue
			}
			w.logger.Info(logging.Internal, logging.Api, "Queue depth", map[logging.ExtraKey]interface{}{
				"queueDepth":     depth,
				"inMemoryQueue": len(w.jobQueue),
				"queueCapacity":  cap(w.jobQueue),
			})
		case <-w.ctx.Done():
			return
		}
	}
}

// digestProcessor periodically processes accumulated digest notifications
func (w *NotificationWorker) digestProcessor() {
	defer w.wg.Done()

	interval := time.Duration(w.config.Digest.Interval) * time.Second
	if interval < 10*time.Second {
		interval = 60 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	w.logger.Debug(logging.General, logging.Startup, "Digest processor started", map[logging.ExtraKey]interface{}{
		"interval": interval.String(),
	})

	for {
		select {
		case <-ticker.C:
			if w.digestService != nil {
				w.digestService.ProcessDueDigests(context.Background())
			}
		case <-w.ctx.Done():
			w.logger.Debug(logging.General, logging.Startup, "Digest processor shutting down", nil)
			return
		}
	}
}

// attemptCleanupProcessor periodically purges expired provider attempt bodies
// and hard-deletes expired attempt rows per the configured retention policy.
// Cleanup is observable via logs and never touches notification records.
//
// NOTE: This is the LEGACY hardcoded cleanup path controlled by env vars
// (PROVIDER_LOGS_*). When a retention policy is enabled through the admin API
// for the "provider_attempts" category, the unified RetentionScheduler takes
// over and this legacy processor should be disabled (set PROVIDER_LOGS_ENABLED=false)
// to avoid both systems running concurrently on the same table.
func (w *NotificationWorker) attemptCleanupProcessor() {
	defer w.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	w.logger.Debug(logging.General, logging.Startup, "Provider attempt retention cleanup started", nil)

	for {
		select {
		case <-ticker.C:
			w.processAttemptCleanup()
		case <-w.ctx.Done():
			w.logger.Debug(logging.General, logging.Startup, "Provider attempt cleanup shutting down", nil)
			return
		}
	}
}

// processAttemptCleanup applies body + metadata retention for attempt logs.
func (w *NotificationWorker) processAttemptCleanup() {
	ctx := context.Background()
	cfg := w.config.ProviderLogs
	if !cfg.Enabled {
		return
	}

	bodyRetention := time.Duration(cfg.BodyRetentionDays) * 24 * time.Hour
	if bodyRetention > 0 {
		purged, err := w.attemptRepo.PurgeBodies(ctx, time.Now().Add(-bodyRetention))
		if err != nil {
			w.logger.Warn(logging.General, logging.Update, "Failed to purge expired attempt bodies", map[logging.ExtraKey]interface{}{
				logging.ExtraKey("error"): err.Error(),
			})
		} else if purged > 0 {
			w.logger.Info(logging.General, logging.Update, "Purged expired provider attempt bodies", map[logging.ExtraKey]interface{}{
				logging.ExtraKey("count"): purged,
			})
		}
	}

	metaRetention := time.Duration(cfg.MetadataRetentionDays) * 24 * time.Hour
	if metaRetention > 0 {
		deleted, err := w.attemptRepo.DeleteExpired(ctx, time.Now().Add(-metaRetention))
		if err != nil {
			w.logger.Warn(logging.General, logging.Update, "Failed to delete expired provider attempts", map[logging.ExtraKey]interface{}{
				logging.ExtraKey("error"): err.Error(),
			})
		} else if deleted > 0 {
			w.logger.Info(logging.General, logging.Update, "Deleted expired provider attempts", map[logging.ExtraKey]interface{}{
				logging.ExtraKey("count"): deleted,
			})
		}
	}
}

// createLog creates a notification log entry
func (w *NotificationWorker) createLog(ctx context.Context, notificationID uuid.UUID, action string, status models.NotificationStatus, message, errorDetails string) {
	log := &models.NotificationLog{
		NotificationID: notificationID,
		Action:         action,
		Status:         status,
		Message:        message,
		ErrorDetails:   errorDetails,
	}

	if err := w.logRepo.Create(ctx, log); err != nil {
		w.logger.Error(logging.General, logging.Insert, "Failed to create notification log", map[logging.ExtraKey]interface{}{
			"notificationId": notificationID,
			"error":          err.Error(),
		})
	}
}
