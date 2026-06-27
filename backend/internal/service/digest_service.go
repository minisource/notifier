package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/repository"
	"github.com/minisource/notifier/internal/worker"
)

// DigestService handles digest/batch notification processing.
// It collects notifications marked as "digested", groups them by user,
// and sends a single digest notification containing the accumulated items.
type DigestService struct {
	logger      logging.Logger
	cfg         *config.Config
	notifRepo   repository.NotificationRepository
	prefRepo    repository.NotificationPreferenceRepository
	smsWorker   worker.SMSHandler
	emailWorker worker.EmailHandler
	worker      *worker.NotificationWorker
}

// NewDigestService creates a new digest service
func NewDigestService(
	cfg *config.Config,
	logger logging.Logger,
	notifRepo repository.NotificationRepository,
	prefRepo repository.NotificationPreferenceRepository,
	smsWorker worker.SMSHandler,
	emailWorker worker.EmailHandler,
	w *worker.NotificationWorker,
) *DigestService {
	return &DigestService{
		logger:      logger,
		cfg:         cfg,
		notifRepo:   notifRepo,
		prefRepo:    prefRepo,
		smsWorker:   smsWorker,
		emailWorker: emailWorker,
		worker:      w,
	}
}

// SetWorker sets the worker reference on the existing DigestService instance.
// This is used during initialization to break the circular dependency
// (DigestService needs Worker, Worker needs DigestProcessor interface).
func (d *DigestService) SetWorker(w *worker.NotificationWorker) {
	d.worker = w
}

// ProcessDueDigests is the main entry point — called periodically by the worker.
// It finds all users with digested notifications, groups by channel, and sends digests.
func (d *DigestService) ProcessDueDigests(ctx context.Context) {
	d.logger.Debug(logging.General, logging.Api, "Processing due digests", nil)

	// Fetch all digested notifications (across all users)
	allDigested, err := d.notifRepo.GetAllDigestedNotifications(ctx, d.cfg.Digest.BatchSize*10)
	if err != nil {
		d.logger.Error(logging.General, logging.Api, "Failed to fetch digested notifications", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return
	}

	if len(allDigested) == 0 {
		return
	}

	// Group by user ID
	byUser := make(map[uuid.UUID][]*models.Notification)
	for _, n := range allDigested {
		byUser[n.UserID] = append(byUser[n.UserID], n)
	}

	d.logger.Info(logging.General, logging.Api, "Found users with digested notifications", map[logging.ExtraKey]interface{}{
		"userCount":       len(byUser),
		"totalItems":      len(allDigested),
	})

	for userID, items := range byUser {
		d.processUserDigests(ctx, userID, items)
	}
}

// processUserDigests processes all digested notifications for a single user
func (d *DigestService) processUserDigests(ctx context.Context, userID uuid.UUID, notifications []*models.Notification) {
	if len(notifications) == 0 {
		return
	}

	d.logger.Debug(logging.General, logging.Api, "Processing user digests", map[logging.ExtraKey]interface{}{
		"userId": userID,
		"count":  len(notifications),
	})

	// Group by notification type (channel)
	byChannel := make(map[models.NotificationType][]*models.Notification)
	for _, n := range notifications {
		byChannel[n.Type] = append(byChannel[n.Type], n)
	}

	for channel, items := range byChannel {
		d.sendDigestForChannel(ctx, userID, channel, items)
	}
}

// sendDigestForChannel sends a digest notification for a specific channel
func (d *DigestService) sendDigestForChannel(ctx context.Context, userID uuid.UUID, channel models.NotificationType, items []*models.Notification) {
	d.logger.Info(logging.General, logging.Api, "Sending digest for user", map[logging.ExtraKey]interface{}{
		"userId":  userID,
		"channel": channel,
		"count":   len(items),
	})

	// Build digest content
	subject, body, _ := d.buildDigestContent(channel, items)
	if subject == "" && body == "" {
		d.logger.Warn(logging.General, logging.Api, "Empty digest content, skipping", map[logging.ExtraKey]interface{}{
			"userId":  userID,
			"channel": channel,
		})
		return
	}

	// Create a single digest notification
	digestNotif := &models.Notification{
		UserID:   userID,
		Type:     channel,
		Status:   models.NotificationStatusPending,
		Priority: models.NotificationPriorityNormal,
		Subject:  subject,
		Body:     body,
		Metadata: fmt.Sprintf(`{"digest":true,"itemCount":%d,"channel":"%s"}`, len(items), channel),
	}

	// Set recipient based on channel using the first matching item
	for _, item := range items {
		if item.RecipientEmail != "" && channel == models.NotificationTypeEmail {
			digestNotif.RecipientEmail = item.RecipientEmail
		}
		if item.RecipientPhone != "" && channel == models.NotificationTypeSMS {
			digestNotif.RecipientPhone = item.RecipientPhone
		}
	}

	// Save the digest notification
	if err := d.notifRepo.Create(ctx, digestNotif); err != nil {
		d.logger.Error(logging.General, logging.Insert, "Failed to create digest notification", map[logging.ExtraKey]interface{}{
			"userId":  userID,
			"channel": channel,
			"error":   err.Error(),
		})
		return
	}

	// Mark original items as sent with sent_at and provider reference
	// MarkAsSent handles the full transition (status + sent_at + provider_msg_id)
	for _, item := range items {
		_ = d.notifRepo.MarkAsSent(ctx, item.ID, fmt.Sprintf("digest-%s", digestNotif.ID.String()[:8]))
	}

	// Enqueue the digest notification for delivery
	if d.worker == nil {
		d.logger.Error(logging.General, logging.Insert, "Worker not set on DigestService, cannot enqueue", map[logging.ExtraKey]interface{}{
			"digestId": digestNotif.ID,
			"userId":   userID,
		})
		return
	}

	if err := d.worker.EnqueueNotification(digestNotif); err != nil {
		d.logger.Error(logging.General, logging.Insert, "Failed to enqueue digest notification", map[logging.ExtraKey]interface{}{
			"digestId": digestNotif.ID,
			"userId":   userID,
			"error":    err.Error(),
		})
		return
	}

	d.logger.Info(logging.General, logging.Api, "Digest sent successfully", map[logging.ExtraKey]interface{}{
		"userId":    userID,
		"channel":   channel,
		"itemCount": len(items),
		"digestId":  digestNotif.ID,
	})
}

// buildDigestContent constructs the subject and body for a digest notification
func (d *DigestService) buildDigestContent(channel models.NotificationType, items []*models.Notification) (subject, body string, isHTML bool) {
	if len(items) == 0 {
		return "", "", false
	}

	switch channel {
	case models.NotificationTypeEmail:
		return d.buildEmailDigest(items)
	case models.NotificationTypeSMS:
		return d.buildSMSDigest(items)
	case models.NotificationTypePush:
		return d.buildPushDigest(items)
	default:
		return d.buildGenericDigest(items)
	}
}

// buildEmailDigest creates an HTML email digest
func (d *DigestService) buildEmailDigest(items []*models.Notification) (string, string, bool) {
	totalCount := len(items)
	if totalCount == 0 {
		return "", "", false
	}

	subject := fmt.Sprintf("You have %d new notification%s", totalCount, pluralSuffix(totalCount))

	now := time.Now().Format("Monday, January 2, 2006")
	var body strings.Builder
	body.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>	<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f7f7f7;">
<div style="background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1);">
<div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 24px; text-align: center;">
	<h1 style="color: white; margin: 0; font-size: 22px;">Notification Digest</h1>
	<p style="color: rgba(255,255,255,0.85); margin: 8px 0 0; font-size: 14px;">%s - %d item%s</p>
</div>
<div style="padding: 24px;">
	<p style="color: #666; font-size: 14px; margin-top: 0;">Here's a summary of your recent notifications:</p>
`, now, totalCount, pluralSuffix(totalCount)))

	for i, item := range items {
		title := item.Subject
		if title == "" {
			title = truncateString(item.Body, d.cfg.Digest.MaxBodyLen)
		}
		summary := truncateString(item.Body, d.cfg.Digest.MaxBodyLen)

		body.WriteString(fmt.Sprintf(`
	<div style="border: 1px solid #eee; border-radius: 8px; padding: 16px; margin-bottom: 12px; background: #fafafa;">
		<div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
			<span style="background: #667eea; color: white; border-radius: 12px; padding: 2px 10px; font-size: 11px; font-weight: 600;">%s</span>
			<span style="color: #999; font-size: 12px;">%s</span>
		</div>
		<h3 style="margin: 0 0 4px; font-size: 15px;">%s</h3>
		<p style="margin: 0; color: #555; font-size: 13px; line-height: 1.4;">%s</p>
	</div>`,
			item.Type,
			item.CreatedAt.Format("Jan 2, 15:04"),
			title,
			summary))

		if i >= d.cfg.Digest.BatchSize {
			remaining := totalCount - i - 1
			if remaining > 0 {
				body.WriteString(fmt.Sprintf(`
	<div style="text-align: center; padding: 16px; color: #999; font-size: 13px;">
		... and %d more notification%s
	</div>`, remaining, pluralSuffix(remaining)))
			}
			break
		}
	}

	body.WriteString(`
</div>
<div style="background: #f0f0f0; padding: 16px 24px; text-align: center; color: #999; font-size: 12px;">
	<p style="margin: 0;">This is an automated digest. Manage preferences in your account settings.</p>
</div>
</div>
</body>
</html>`)

	return subject, body.String(), true
}

// buildSMSDigest creates a plain-text SMS digest
func (d *DigestService) buildSMSDigest(items []*models.Notification) (string, string, bool) {
	totalCount := len(items)
	if totalCount == 0 {
		return "", "", false
	}

	subject := fmt.Sprintf("Digest: %d notification%s", totalCount, pluralSuffix(totalCount))

	var body strings.Builder
	body.WriteString(fmt.Sprintf("Digest - %d notification%s:\n", totalCount, pluralSuffix(totalCount)))

	for i, item := range items {
		if i >= 3 {
			remaining := totalCount - i
			body.WriteString(fmt.Sprintf("+%d more", remaining))
			break
		}
		title := item.Subject
		if title == "" {
			title = truncateString(item.Body, 80)
		}
		body.WriteString(fmt.Sprintf("\n- %s", title))
	}

	body.WriteString("\n\nReply to manage preferences")
	return subject, body.String(), false
}

// buildPushDigest creates a short push notification digest
func (d *DigestService) buildPushDigest(items []*models.Notification) (string, string, bool) {
	totalCount := len(items)
	if totalCount == 0 {
		return "", "", false
	}

	subject := fmt.Sprintf("%d new notification%s", totalCount, pluralSuffix(totalCount))

	typeCounts := make(map[models.NotificationType]int)
	for _, item := range items {
		typeCounts[item.Type]++
	}
	var parts []string
	for t, c := range typeCounts {
		parts = append(parts, fmt.Sprintf("%d %s", c, t))
	}
	body := strings.Join(parts, ", ")

	return subject, body, false
}

// buildGenericDigest creates a generic plain-text digest (fallback)
func (d *DigestService) buildGenericDigest(items []*models.Notification) (string, string, bool) {
	totalCount := len(items)
	if totalCount == 0 {
		return "", "", false
	}

	subject := fmt.Sprintf("Digest: %d notification%s", totalCount, pluralSuffix(totalCount))
	var body strings.Builder
	body.WriteString(fmt.Sprintf("You have %d notification%s:\n\n", totalCount, pluralSuffix(totalCount)))

	for i, item := range items {
		if i >= d.cfg.Digest.BatchSize {
			break
		}
		title := item.Subject
		if title == "" {
			title = truncateString(item.Body, d.cfg.Digest.MaxBodyLen)
		}
		body.WriteString(fmt.Sprintf("[%s] %s\n", item.Type, title))
	}

	return subject, body.String(), false
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// pluralSuffix returns "s" for plural or "" for singular
func pluralSuffix(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
