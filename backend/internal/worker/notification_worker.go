package worker

import (
	"context"
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
	smsHandler    SMSHandler
	emailHandler  EmailHandler
	pushHandler   PushHandler
	digestService DigestProcessor
}

// DigestProcessor interface for processing digest notifications
type DigestProcessor interface {
	ProcessDueDigests(ctx context.Context)
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
	smsHandler SMSHandler,
	emailHandler EmailHandler,
	pushHandler PushHandler,
	digestProcessor DigestProcessor,
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
		smsHandler:    smsHandler,
		emailHandler:  emailHandler,
		pushHandler:   pushHandler,
		digestService: digestProcessor,
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

	// Create success log
	w.createLog(ctx, notification.ID, "sent", models.NotificationStatusSent, "Notification sent successfully", "")
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
