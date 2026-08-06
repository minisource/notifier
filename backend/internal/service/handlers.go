// Package service provides production-ready notification handling implementations
// SMS, Email, and Push notifications are sent via configurable providers loaded from database

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/common"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/attemptlog"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/platform/email"
	"github.com/minisource/notifier/internal/platform/push"
	"github.com/minisource/notifier/internal/platform/sms"
	smsplatforms "github.com/minisource/notifier/internal/platform/sms/platforms"
	providerpkg "github.com/minisource/notifier/internal/provider"
	"github.com/minisource/notifier/internal/worker"
)

// Helper function to get map keys for logging
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// SMSHandlerAdapter handles SMS notifications using database-configured providers
type SMSHandlerAdapter struct {
	service *NotificationService
}

// NewSMSHandlerAdapter creates SMS handler
func NewSMSHandlerAdapter(service *NotificationService) *SMSHandlerAdapter {
	return &SMSHandlerAdapter{service: service}
}

// SendSMS sends SMS using the configured provider from database settings
// Template-based flow:
// 1. Parse metadata for template key and data
// 2. Load SMS provider config from database
// 3. Look up template mapping in sms_templates table
// 4. For lookup-based providers (Kavenegar): use provider_template + mapped tokens
// 5. For text-based providers: use message_template with placeholder replacement
func (a *SMSHandlerAdapter) SendSMS(ctx context.Context, notification *models.Notification) (string, error) {
	// FINAL authoritative pause gate at the provider boundary: no outbound
	// SMS may start while the global delivery pause is active. This covers
	// worker, sync, and digest send paths alike.
	if a.service.IsDeliveryPaused(ctx) {
		return "", worker.ErrDeliveryPaused
	}

	// Normalize phone number to E.164 format
	normalizedPhone := common.NormalizeIranPhone(notification.RecipientPhone)

	a.service.logger.Info(logging.General, logging.Api,
		"Sending SMS notification",
		map[logging.ExtraKey]interface{}{
			"notificationID":  notification.ID,
			"phone":           notification.RecipientPhone,
			"normalizedPhone": normalizedPhone,
		})

	// Parse metadata for template key and data
	// Metadata format: {"template": "verify", "data": {"code": "123456"}, ...}
	var templateKey string
	data := make(map[string]string)

	a.service.logger.Info(logging.General, logging.Api, "========== RAW METADATA BEFORE PARSING ==========", map[logging.ExtraKey]interface{}{
		"notificationID":    notification.ID,
		"metadata_raw":      notification.Metadata,
		"metadata_len":      len(notification.Metadata),
		"metadata_is_empty": notification.Metadata == "",
	})

	if notification.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(notification.Metadata), &metadata); err == nil {
			a.service.logger.Info(logging.General, logging.Api, "========== METADATA PARSED SUCCESSFULLY ==========", map[logging.ExtraKey]interface{}{
				"notificationID": notification.ID,
				"metadata_keys":  getKeys(metadata),
				"metadata_full":  metadata,
			})

			// Extract template key
			if template, ok := metadata["template"].(string); ok && template != "" {
				templateKey = template
			}

			// Extract data map
			a.service.logger.Info(logging.General, logging.Api, "========== EXTRACTING DATA FIELD ==========", map[logging.ExtraKey]interface{}{
				"notificationID":   notification.ID,
				"has_data_key":     metadata["data"] != nil,
				"data_field_type":  fmt.Sprintf("%T", metadata["data"]),
				"data_field_value": metadata["data"],
			})

			if dataMap, ok := metadata["data"].(map[string]interface{}); ok {
				a.service.logger.Info(logging.General, logging.Api, "========== DATA FIELD CAST SUCCESSFUL ==========", map[logging.ExtraKey]interface{}{
					"notificationID": notification.ID,
					"dataMap_len":    len(dataMap),
					"dataMap":        dataMap,
				})

				for k, v := range dataMap {
					if str, ok := v.(string); ok {
						data[k] = str
					}
				}
			}
			// Legacy support: extract token values directly from metadata
			for _, key := range []string{"token", "code", "token2", "token3", "token10", "token20"} {
				if val, ok := metadata[key].(string); ok && val != "" {
					data[key] = val
				}
			}
		}
	}

	// If no template key was extracted from metadata but the notification
	// explicitly references a template, resolve its key from the unified
	// template. This makes template-based sends work when the caller only
	// passes templateId + variables (no metadata), which is the canonical
	// create-notification contract.
	if templateKey == "" && notification.TemplateID != nil {
		tpl, tErr := a.service.templateRepo.GetByID(ctx, *notification.TemplateID)
		if tErr == nil && tpl != nil {
			if tpl.Key != "" {
				templateKey = tpl.Key
			} else if tpl.Name != "" {
				templateKey = tpl.Name
			}
		}
	}

	// If no data provided, use body as the code/token. Only for free-text
	// sends WITHOUT a template — when a template is used the body holds
	// placeholders (e.g. "Your verification code is: {{code}}"), not the
	// actual token value, so sending it to a lookup endpoint would fail.
	if len(data) == 0 && templateKey == "" && notification.Body != "" {
		data["code"] = notification.Body
		data["token"] = notification.Body
	}

	a.service.logger.Info(logging.General, logging.Api,
		"SMS metadata parsed",
		map[logging.ExtraKey]interface{}{
			"notificationID": notification.ID,
			"templateKey":    templateKey,
			"data":           data,
			"metadata":       notification.Metadata,
		})

	// Collect provider candidates: configured providers table first (failover
	// order = is_default DESC, priority ASC, tenant-scoped), falling back to the
	// legacy settings-based config when no providers are configured.
	configs, err := a.smsProviderCandidates(ctx, notification)
	if err != nil {
		return "", err
	}

	var lastSendErr error
	var parentAttemptID *uuid.UUID
	rec := a.service.GetProviderAttemptRecorder()
	for idx, config := range configs {
		messageID, attempt, err := a.sendSMSViaProvider(ctx, notification, config, templateKey, data, normalizedPhone, idx, parentAttemptID)
		if err == nil {
			return messageID, nil
		}
		lastSendErr = err
		// Link the next fallback attempt to this failed one and record the event.
		if attempt != nil {
			parentAttemptID = &attempt.ID
			if idx+1 < len(configs) {
				rec.RecordFallbackScheduled(ctx, attempt, configs[idx+1].Provider)
			}
		}
		a.service.logger.Warn(logging.General, logging.Api,
			"SMS provider failed, trying next provider",
			map[logging.ExtraKey]interface{}{
				"notificationID": notification.ID,
				"provider":       config.Provider,
				"error":          err.Error(),
			})
	}

	if lastSendErr != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"All SMS providers failed",
			map[logging.ExtraKey]interface{}{
				"notificationID": notification.ID,
				"error":          lastSendErr.Error(),
			})
		return "", fmt.Errorf("failed to send SMS via all providers: %w", lastSendErr)
	}

	return "", fmt.Errorf("failed to send SMS: no provider attempted")
}

// smsProviderCandidates returns the ordered list of SMS provider configs to
// attempt for a notification. Providers from the providers table (tenant-scoped,
// enabled only, ordered by is_default DESC then priority ASC) are preferred;
// if none are configured it falls back to the legacy settings-based config.
func (a *SMSHandlerAdapter) smsProviderCandidates(ctx context.Context, notification *models.Notification) ([]*sms.ProviderConfig, error) {
	// Explicit provider selection: use ONLY the chosen provider, no failover.
	if notification.ProviderID != nil {
		p, err := a.service.providerRepo.GetByID(ctx, *notification.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("failed to load selected provider: %w", err)
		}
		if p == nil {
			return nil, fmt.Errorf("selected provider not found")
		}
		if !p.IsEnabled || p.Status == models.ProviderStatusDisabled {
			return nil, fmt.Errorf("selected provider %q is disabled", p.Name)
		}
		cfg, err := providerConfigToSMS(p)
		if err != nil {
			return nil, fmt.Errorf("selected provider %q has invalid config: %w", p.Name, err)
		}
		return []*sms.ProviderConfig{cfg}, nil
	}

	providers, err := a.service.providerRepo.List(ctx, "sms", notification.TenantID)
	if err != nil {
		a.service.logger.Warn(logging.General, logging.Select,
			"Failed to list SMS providers from table, falling back to settings",
			map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
	} else if len(providers) > 0 {
		configs := make([]*sms.ProviderConfig, 0, len(providers))
		for _, p := range providers {
			if !p.IsEnabled || p.Status == models.ProviderStatusDisabled {
				continue
			}
			cfg, err := providerConfigToSMS(p)
			if err != nil {
				a.service.logger.Warn(logging.General, logging.Api,
					"Skipping SMS provider with invalid config",
					map[logging.ExtraKey]interface{}{
						"provider": p.Name,
						"error":    err.Error(),
					})
				continue
			}
			configs = append(configs, cfg)
		}
		if len(configs) > 0 {
			return configs, nil
		}
	}

	// Legacy fallback: single provider from settings table
	setting, err := a.service.settingRepo.GetByKey(ctx, models.SettingKeySMSProviders)
	if err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to load SMS provider config from database",
			map[logging.ExtraKey]interface{}{
				"error":      err.Error(),
				"settingKey": models.SettingKeySMSProviders,
			})
		return nil, fmt.Errorf("SMS provider not configured: %w", err)
	}

	config, err := sms.ParseProviderConfig(setting.Value)
	if err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to parse SMS provider config",
			map[logging.ExtraKey]interface{}{
				"error":        err.Error(),
				"settingValue": setting.Value,
			})
		return nil, fmt.Errorf("invalid SMS provider config: %w", err)
	}

	return []*sms.ProviderConfig{config}, nil
}

// sendSMSViaProvider sends the SMS through a single provider config (template
// lookup + params + client + send). Used by the failover loop in SendSMS.
func (a *SMSHandlerAdapter) sendSMSViaProvider(ctx context.Context, notification *models.Notification, config *sms.ProviderConfig, templateKey string, data map[string]string, normalizedPhone string, fallbackSequence int, parentAttemptID *uuid.UUID) (string, *models.ProviderAttempt, error) {
	// Record the provider handling this notification (persisted by the worker on
	// success or failure) so per-provider stats are accurate.
	notification.Provider = config.Provider

	// Prepare message parameters
	params := make(map[string]string)

	// Look up SMS template from database if template key is provided
	if templateKey != "" {
		// 1. Try resolving using the new unified NotificationTemplate
		var unifiedTemplate *models.NotificationTemplate
		var utErr error
		if notification.TemplateID != nil {
			unifiedTemplate, utErr = a.service.templateRepo.GetByID(ctx, *notification.TemplateID)
		} else {
			unifiedTemplate, utErr = a.service.templateRepo.GetByKey(ctx, templateKey)
		}

		if utErr == nil && unifiedTemplate != nil {
			// Find provider template key
			providerTemplateKey := ""
			providerTemplates := unifiedTemplate.ParseProviderTemplates()
			var fallbackKey string
			for _, pt := range providerTemplates {
				if pt.Provider == config.Provider {
					if pt.TenantID != "" && notification.TenantID != nil && pt.TenantID == notification.TenantID.String() {
						providerTemplateKey = pt.TemplateKey
						break
					}
					if pt.TenantID == "" {
						fallbackKey = pt.TemplateKey
					}
				}
			}
			if providerTemplateKey == "" {
				providerTemplateKey = fallbackKey
			}

			// Fallback to name if not specified
			if providerTemplateKey == "" {
				providerTemplateKey = unifiedTemplate.Name
				if providerTemplateKey == "" {
					providerTemplateKey = unifiedTemplate.Key
				}
			}

			if providerTemplateKey != "" {
				params["template"] = providerTemplateKey
				// Copy all variables from data to params
				for k, v := range data {
					params[k] = v
				}

				// Keep lookups happy: if there is no "token" but "code" exists, copy code to token
				if _, ok := params["token"]; !ok {
					if codeVal, ok := data["code"]; ok {
						params["token"] = codeVal
					}
				}

				a.service.logger.Info(logging.General, logging.Api,
					"Resolved SMS provider template using unified NotificationTemplate",
					map[logging.ExtraKey]interface{}{
						"templateKey":         templateKey,
						"providerTemplateKey": providerTemplateKey,
						"params":              params,
					})
			}
		}

		// 2. If not resolved via unified template, fallback to the legacy SMSTemplate repo
		if _, exists := params["template"]; !exists {
			smsTemplate, err := a.service.smsTemplateRepo.GetByKeyAndProvider(ctx, templateKey, config.Provider, nil)
			if err != nil {
				a.service.logger.Warn(logging.General, logging.Api,
					"SMS template not found in database, using template key as-is",
					map[logging.ExtraKey]interface{}{
						"templateKey": templateKey,
						"provider":    config.Provider,
						"error":       err.Error(),
					})
				// Use template key as-is (fallback behavior)
				params["template"] = templateKey
				// Copy data directly to params
				for k, v := range data {
					params[k] = v
				}
			} else {
				// Template found - determine send type
				if smsTemplate.ProviderTemplate != "" {
					// Lookup-based provider (e.g., Kavenegar)
					params["template"] = smsTemplate.ProviderTemplate
					// Map tokens according to template mapping
					mappedTokens := smsTemplate.MapTokens(data)
					for k, v := range mappedTokens {
						params[k] = v
					}
					a.service.logger.Info(logging.General, logging.Api,
						"Using lookup-based SMS template",
						map[logging.ExtraKey]interface{}{
							"templateKey":      templateKey,
							"providerTemplate": smsTemplate.ProviderTemplate,
							"inputData":        data,
							"mappedTokens":     mappedTokens,
							"params":           params,
						})
				} else if smsTemplate.MessageTemplate != "" {
					// Text-based provider - replace placeholders in message template
					message := smsTemplate.MessageTemplate
					for k, v := range data {
						message = replacePlaceholder(message, k, v)
					}
					params["message"] = message
					params["body"] = message
					a.service.logger.Debug(logging.General, logging.Api,
						"Using text-based SMS template",
						map[logging.ExtraKey]interface{}{
							"templateKey": templateKey,
							"message":     message[:min(50, len(message))] + "...",
						})
				}
			}
		}
	} else {
		// No template - use body directly
		params["message"] = notification.Body
		params["body"] = notification.Body
		params["token"] = notification.Body
		params["code"] = notification.Body
	}

	// ---- Provider Request Lifecycle Logging (durable attempt) ----
	// One attempt row per provider call. Request metadata is sanitized before
	// persistence: sensitive params (code/token/apiKey) are redacted, the
	// recipient is never stored, and only a bounded preview + hash of the
	// message content is captured. The attempt is opened BEFORE client
	// creation so configuration failures are also recorded.
	opts := attemptlog.DefaultRedactionOptions()
	rawBody, _ := json.Marshal(params)
	bodySan, _, _, capturedSize := attemptlog.SanitizeBody(string(rawBody), opts)
	contentHash := attemptlog.ContentHash(notification.Body)
	preview := attemptlog.TruncatePreview(notification.Body, opts.MaxPreviewChars)
	rec := a.service.GetProviderAttemptRecorder()
	attempt := rec.Start(ctx, attemptlog.StartInput{
		NotificationID:      notification.ID,
		TenantID:            notification.TenantID,
		ParentAttemptID:     parentAttemptID,
		Channel:             "sms",
		Provider:            config.Provider,
		AttemptNumber:       notification.RetryCount + 1,
		FallbackSequence:    fallbackSequence,
		RequestMethod:       "POST",
		RequestURLSanitized: "sms://" + config.Provider + "/send",
		RequestHeadersSanitized: map[string]string{},
		RequestBodySanitized:    bodySan,
		RequestSizeBytes:        capturedSize,
		ContentHash:             contentHash,
		BodyPreview:             preview,
		RecipientMasked:         attemptlog.MaskPhone(normalizedPhone),
		CorrelationID:           notification.ID.String(),
	})
	startedAt := time.Now()
	rec.MarkSending(ctx, attempt)

	// Create SMS client
	client, err := sms.NewClientFromConfig(config)
	if err != nil {
		kind, code := attemptlog.ClassifyErrorText(err.Error())
		rec.Finish(ctx, attempt, attemptlog.FinishInput{
			Status:                 attemptlog.StatusRejected,
			NormalizedErrorKind:    string(kind),
			NormalizedErrorCode:    code,
			NormalizedErrorMessage: err.Error(),
			Retryable:              false,
			StartedAt:              startedAt,
		})
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to create SMS client from config",
			map[logging.ExtraKey]interface{}{
				"error":    err.Error(),
				"provider": config.Provider,
			})
			return "", attempt, fmt.Errorf("failed to create SMS client: %w", err)
	}

	// For Twilio, we need sender number as first target
	var targets []string
	if config.Provider == "twilio" && config.SenderID != "" {
		targets = append(targets, config.SenderID)
	}
	targets = append(targets, normalizedPhone)

	// Replace the placeholder request metadata with the REAL outbound request
	// (URL + form body) once the client knows its exact endpoint. The adapter
	// returns an already-sanitized description — API key redacted, recipient
	// masked, OTP tokens redacted — so the attempt log shows exactly what the
	// provider receives, not the internal params map.
	if describer, ok := client.(smsplatforms.RequestDescriber); ok {
		method, reqURL, reqBody := describer.DescribeRequest(params, targets...)
		if reqURL != "" {
			bodySan, _, _, reqSize := attemptlog.SanitizeBody(reqBody, opts)
			rec.UpdateRequest(ctx, attempt, method, attemptlog.SanitizeURL(reqURL), bodySan, reqSize)
		}
	}

	// Send SMS
	if err := client.SendMessage(params, targets...); err != nil {
		kind, code := attemptlog.ClassifyErrorText(err.Error())
		rec.Finish(ctx, attempt, attemptlog.FinishInput{
			Status:                 attemptlog.StatusFailed,
			NormalizedErrorKind:    string(kind),
			NormalizedErrorCode:    code,
			NormalizedErrorMessage: err.Error(),
			Retryable:              true,
			StartedAt:              startedAt,
		})
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to send SMS",
			map[logging.ExtraKey]interface{}{
				"notificationID": notification.ID,
				"provider":       config.Provider,
				"error":          err.Error(),
			})
		return "", attempt, fmt.Errorf("failed to send SMS: %w", err)
	}

	messageID := fmt.Sprintf("sms-%s", notification.ID.String()[:8])
	rec.Finish(ctx, attempt, attemptlog.FinishInput{
		Status:            attemptlog.StatusAccepted,
		ProviderStatus:    "accepted",
		ProviderMessageID: messageID,
		StartedAt:         startedAt,
	})
	a.service.logger.Info(logging.General, logging.Api,
		"SMS sent successfully",
		map[logging.ExtraKey]interface{}{
			"notificationID": notification.ID,
			"messageID":      messageID,
			"provider":       config.Provider,
		})

	return messageID, attempt, nil
}

// providerConfigToSMS builds an sms.ProviderConfig from a providers-table row.
// It also injects the provider type from the row's Type field when the config
// doesn't already carry a "provider" key.
func providerConfigToSMS(p *models.Provider) (*sms.ProviderConfig, error) {
	raw, err := providerpkg.MergeProviderConfigJSON(p)
	if err != nil {
		return nil, err
	}
	cfg, err := sms.ParseProviderConfig(raw)
	if err != nil {
		return nil, err
	}
	if cfg.Provider == "" && p.Type != "" {
		cfg.Provider = p.Type
	}
	return cfg, nil
}

// providerConfigToEmail builds an email.ProviderConfig from a providers-table row.
func providerConfigToEmail(p *models.Provider) (*email.ProviderConfig, error) {
	raw, err := providerpkg.MergeProviderConfigJSON(p)
	if err != nil {
		return nil, err
	}
	cfg, err := email.ParseProviderConfig(raw)
	if err != nil {
		return nil, err
	}
	if cfg.Provider == "" && p.Type != "" {
		cfg.Provider = p.Type
	}
	return cfg, nil
}

// providerConfigToPush builds a push.ProviderConfig from a providers-table row.
func providerConfigToPush(p *models.Provider) (*push.ProviderConfig, error) {
	raw, err := providerpkg.MergeProviderConfigJSON(p)
	if err != nil {
		return nil, err
	}
	cfg, err := push.ParseProviderConfig(raw)
	if err != nil {
		return nil, err
	}
	if cfg.Provider == "" && p.Type != "" {
		cfg.Provider = p.Type
	}
	return cfg, nil
}

// replacePlaceholder replaces {{key}} with value in template
func replacePlaceholder(template, key, value string) string {
	placeholder := "{{" + key + "}}"
	return strings.ReplaceAll(template, placeholder, value)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// EmailHandlerAdapter handles email notifications using database-configured providers
type EmailHandlerAdapter struct {
	service *NotificationService
}

// NewEmailHandlerAdapter creates email handler
func NewEmailHandlerAdapter(service *NotificationService) *EmailHandlerAdapter {
	return &EmailHandlerAdapter{service: service}
}

// SendEmail sends email using the configured provider from database settings
func (a *EmailHandlerAdapter) SendEmail(ctx context.Context, notification *models.Notification) (string, error) {
	// FINAL authoritative pause gate at the provider boundary.
	if a.service.IsDeliveryPaused(ctx) {
		return "", worker.ErrDeliveryPaused
	}

	a.service.logger.Info(logging.General, logging.Api,
		"Sending email notification",
		map[logging.ExtraKey]interface{}{
			"notificationID": notification.ID,
			"email":          notification.RecipientEmail,
		})

	// Collect email provider candidates: providers table first (failover order),
	// then legacy settings-based config.
	configs, err := a.emailProviderCandidates(ctx, notification)
	if err != nil {
		return "", err
	}

	// Determine if content is HTML (check for common HTML tags)
	isHTML := len(notification.Body) > 0 && (notification.Body[0] == '<' ||
		len(notification.Body) > 5 && notification.Body[:5] == "<!DOC")

	var lastSendErr error
	for idx, config := range configs {
		// Record the provider handling this notification (persisted by the worker).
		notification.Provider = config.Provider

		// ---- Provider Request Lifecycle Logging (durable attempt) ----
		opts := attemptlog.DefaultRedactionOptions()
		contentHash := attemptlog.ContentHash(notification.Body)
		preview := attemptlog.TruncatePreview(notification.Body, opts.MaxPreviewChars)
		rec := a.service.GetProviderAttemptRecorder()
		attempt := rec.Start(ctx, attemptlog.StartInput{
			NotificationID:      notification.ID,
			TenantID:            notification.TenantID,
			Channel:             "email",
			Provider:            config.Provider,
			AttemptNumber:       notification.RetryCount + 1,
			FallbackSequence:    idx,
			RequestMethod:       "SMTP",
			RequestURLSanitized: "smtp://" + config.Provider + "/send",
			RequestHeadersSanitized: map[string]string{},
			RequestBodySanitized:    "",
			RequestSizeBytes:        0,
			ContentHash:             contentHash,
			BodyPreview:             preview,
			RecipientMasked:         attemptlog.MaskEmail(notification.RecipientEmail),
			CorrelationID:           notification.ID.String(),
		})
		startedAt := time.Now()
		rec.MarkSending(ctx, attempt)

		// Create email client
		client, err := email.NewClientFromConfig(config)
		if err != nil {
			kind, code := attemptlog.ClassifyErrorText(err.Error())
			rec.Finish(ctx, attempt, attemptlog.FinishInput{
				Status:                 attemptlog.StatusRejected,
				NormalizedErrorKind:    string(kind),
				NormalizedErrorCode:    code,
				NormalizedErrorMessage: err.Error(),
				Retryable:              false,
				StartedAt:              startedAt,
			})
			lastSendErr = fmt.Errorf("failed to create email client: %w", err)
			a.service.logger.Warn(logging.General, logging.Api,
				"Email provider skipped (client creation failed), trying next",
				map[logging.ExtraKey]interface{}{
					"provider": config.Provider,
					"error":    err.Error(),
				})
			continue
		}

		// Send email
		if err := client.SendEmail(notification.RecipientEmail, notification.Subject, notification.Body, isHTML); err != nil {
			kind, code := attemptlog.ClassifyErrorText(err.Error())
			rec.Finish(ctx, attempt, attemptlog.FinishInput{
				Status:                 attemptlog.StatusFailed,
				NormalizedErrorKind:    string(kind),
				NormalizedErrorCode:    code,
				NormalizedErrorMessage: err.Error(),
				Retryable:              true,
				StartedAt:              startedAt,
			})
			lastSendErr = fmt.Errorf("failed to send email: %w", err)
			a.service.logger.Warn(logging.General, logging.Api,
				"Email provider failed, trying next provider",
				map[logging.ExtraKey]interface{}{
					"notificationID": notification.ID,
					"provider":       config.Provider,
					"error":          err.Error(),
				})
			continue
		}

		messageID := fmt.Sprintf("email-%s", notification.ID.String()[:8])
		rec.Finish(ctx, attempt, attemptlog.FinishInput{
			Status:            attemptlog.StatusAccepted,
			ProviderStatus:    "accepted",
			ProviderMessageID: messageID,
			StartedAt:         startedAt,
		})
		a.service.logger.Info(logging.General, logging.Api,
			"Email sent successfully",
			map[logging.ExtraKey]interface{}{
				"notificationID": notification.ID,
				"messageID":      messageID,
				"provider":       config.Provider,
			})

		return messageID, nil
	}

	if lastSendErr != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"All email providers failed",
			map[logging.ExtraKey]interface{}{
				"notificationID": notification.ID,
				"error":          lastSendErr.Error(),
			})
		return "", lastSendErr
	}

	return "", fmt.Errorf("failed to send email: no provider attempted")
}

// emailProviderCandidates returns the ordered list of email provider configs to
// attempt (providers table first, then legacy settings config).
func (a *EmailHandlerAdapter) emailProviderCandidates(ctx context.Context, notification *models.Notification) ([]*email.ProviderConfig, error) {
	// Explicit provider selection: use ONLY the chosen provider, no failover.
	if notification.ProviderID != nil {
		p, err := a.service.providerRepo.GetByID(ctx, *notification.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("failed to load selected provider: %w", err)
		}
		if p == nil {
			return nil, fmt.Errorf("selected provider not found")
		}
		if !p.IsEnabled || p.Status == models.ProviderStatusDisabled {
			return nil, fmt.Errorf("selected provider %q is disabled", p.Name)
		}
		cfg, err := providerConfigToEmail(p)
		if err != nil {
			return nil, fmt.Errorf("selected provider %q has invalid config: %w", p.Name, err)
		}
		return []*email.ProviderConfig{cfg}, nil
	}

	providers, err := a.service.providerRepo.List(ctx, "email", notification.TenantID)
	if err != nil {
		a.service.logger.Warn(logging.General, logging.Select,
			"Failed to list email providers from table, falling back to settings",
			map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
	} else if len(providers) > 0 {
		configs := make([]*email.ProviderConfig, 0, len(providers))
		for _, p := range providers {
			if !p.IsEnabled || p.Status == models.ProviderStatusDisabled {
				continue
			}
			cfg, err := providerConfigToEmail(p)
			if err != nil {
				a.service.logger.Warn(logging.General, logging.Api,
					"Skipping email provider with invalid config",
					map[logging.ExtraKey]interface{}{
						"provider": p.Name,
						"error":    err.Error(),
					})
				continue
			}
			configs = append(configs, cfg)
		}
		if len(configs) > 0 {
			return configs, nil
		}
	}

	// Legacy fallback: single provider from settings table
	setting, err := a.service.settingRepo.GetByKey(ctx, models.SettingKeyEmailProviders)
	if err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to load email provider config",
			map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
		return nil, fmt.Errorf("email provider not configured: %w", err)
	}

	config, err := email.ParseProviderConfig(setting.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid email provider config: %w", err)
	}

	return []*email.ProviderConfig{config}, nil
}

// PushHandlerAdapter handles push notifications using database-configured providers
type PushHandlerAdapter struct {
	service *NotificationService
}

// NewPushHandlerAdapter creates push handler
func NewPushHandlerAdapter(service *NotificationService) *PushHandlerAdapter {
	return &PushHandlerAdapter{service: service}
}

// Push provider setting key
const SettingKeyPushProviders = "push.providers"

// SendPush sends push notification using the configured provider from database settings
func (a *PushHandlerAdapter) SendPush(ctx context.Context, notification *models.Notification) (string, error) {
	// FINAL authoritative pause gate at the provider boundary.
	if a.service.IsDeliveryPaused(ctx) {
		return "", worker.ErrDeliveryPaused
	}

	a.service.logger.Info(logging.General, logging.Api,
		"Sending push notification",
		map[logging.ExtraKey]interface{}{
			"notificationID": notification.ID,
			"recipientID":    notification.RecipientID,
		})

	// Collect push provider candidates: providers table first (failover order),
	// then legacy settings-based config.
	configs, err := a.pushProviderCandidates(ctx, notification)
	if err != nil {
		return "", err
	}

	// Parse metadata JSON to get device token and additional data
	var metadata map[string]interface{}
	if notification.Metadata != "" {
		if err := json.Unmarshal([]byte(notification.Metadata), &metadata); err != nil {
			return "", fmt.Errorf("failed to parse notification metadata: %w", err)
		}
	}

	// Device token should be stored in notification metadata
	deviceToken := ""
	if metadata != nil {
		if token, ok := metadata["deviceToken"].(string); ok {
			deviceToken = token
		}
	}

	if deviceToken == "" {
		return "", fmt.Errorf("push device token is required: provide a deviceToken in notification metadata")
	}

	// Prepare data payload from metadata
	data := make(map[string]string)
	for k, v := range metadata {
		if str, ok := v.(string); ok && k != "deviceToken" {
			data[k] = str
		}
	}
	data["notificationId"] = notification.ID.String()

	var lastSendErr error
	for idx, config := range configs {
		// Record the provider handling this notification (persisted by the worker).
		notification.Provider = config.Provider

		// ---- Provider Request Lifecycle Logging (durable attempt) ----
		opts := attemptlog.DefaultRedactionOptions()
		contentHash := attemptlog.ContentHash(notification.Body)
		preview := attemptlog.TruncatePreview(notification.Body, opts.MaxPreviewChars)
		rec := a.service.GetProviderAttemptRecorder()
		attempt := rec.Start(ctx, attemptlog.StartInput{
			NotificationID:      notification.ID,
			TenantID:            notification.TenantID,
			Channel:             "push",
			Provider:            config.Provider,
			AttemptNumber:       notification.RetryCount + 1,
			FallbackSequence:    idx,
			RequestMethod:       "POST",
			RequestURLSanitized: "push://" + config.Provider + "/send",
			RequestHeadersSanitized: map[string]string{},
			RequestBodySanitized:    "",
			RequestSizeBytes:        0,
			ContentHash:             contentHash,
			BodyPreview:             preview,
			RecipientMasked:         attemptlog.MaskRecipient(notification.RecipientID, "push"),
			CorrelationID:           notification.ID.String(),
		})
		startedAt := time.Now()
		rec.MarkSending(ctx, attempt)

		// Create push client
		client, err := push.NewClientFromConfig(config)
		if err != nil {
			kind, code := attemptlog.ClassifyErrorText(err.Error())
			rec.Finish(ctx, attempt, attemptlog.FinishInput{
				Status:                 attemptlog.StatusRejected,
				NormalizedErrorKind:    string(kind),
				NormalizedErrorCode:    code,
				NormalizedErrorMessage: err.Error(),
				Retryable:              false,
				StartedAt:              startedAt,
			})
			lastSendErr = fmt.Errorf("failed to create push client: %w", err)
			a.service.logger.Warn(logging.General, logging.Api,
				"Push provider skipped (client creation failed), trying next",
				map[logging.ExtraKey]interface{}{
					"provider": config.Provider,
					"error":    err.Error(),
				})
			continue
		}

		// Send push notification
		if err := client.SendPush(deviceToken, notification.Subject, notification.Body, data); err != nil {
			kind, code := attemptlog.ClassifyErrorText(err.Error())
			rec.Finish(ctx, attempt, attemptlog.FinishInput{
				Status:                 attemptlog.StatusFailed,
				NormalizedErrorKind:    string(kind),
				NormalizedErrorCode:    code,
				NormalizedErrorMessage: err.Error(),
				Retryable:              true,
				StartedAt:              startedAt,
			})
			lastSendErr = fmt.Errorf("failed to send push notification: %w", err)
			a.service.logger.Warn(logging.General, logging.Api,
				"Push provider failed, trying next provider",
				map[logging.ExtraKey]interface{}{
					"notificationID": notification.ID,
					"provider":       config.Provider,
					"error":          err.Error(),
				})
			continue
		}

		messageID := fmt.Sprintf("push-%s", notification.ID.String()[:8])
		rec.Finish(ctx, attempt, attemptlog.FinishInput{
			Status:            attemptlog.StatusAccepted,
			ProviderStatus:    "accepted",
			ProviderMessageID: messageID,
			StartedAt:         startedAt,
		})
		a.service.logger.Info(logging.General, logging.Api,
			"Push notification sent successfully",
			map[logging.ExtraKey]interface{}{
				"notificationID": notification.ID,
				"messageID":      messageID,
				"provider":       config.Provider,
			})

		return messageID, nil
	}

	if lastSendErr != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"All push providers failed",
			map[logging.ExtraKey]interface{}{
				"notificationID": notification.ID,
				"error":          lastSendErr.Error(),
			})
		return "", lastSendErr
	}

	return "", fmt.Errorf("failed to send push notification: no provider attempted")
}

// pushProviderCandidates returns the ordered list of push provider configs to
// attempt (providers table first, then legacy settings config).
func (a *PushHandlerAdapter) pushProviderCandidates(ctx context.Context, notification *models.Notification) ([]*push.ProviderConfig, error) {
	// Explicit provider selection: use ONLY the chosen provider, no failover.
	if notification.ProviderID != nil {
		p, err := a.service.providerRepo.GetByID(ctx, *notification.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("failed to load selected provider: %w", err)
		}
		if p == nil {
			return nil, fmt.Errorf("selected provider not found")
		}
		if !p.IsEnabled || p.Status == models.ProviderStatusDisabled {
			return nil, fmt.Errorf("selected provider %q is disabled", p.Name)
		}
		cfg, err := providerConfigToPush(p)
		if err != nil {
			return nil, fmt.Errorf("selected provider %q has invalid config: %w", p.Name, err)
		}
		return []*push.ProviderConfig{cfg}, nil
	}

	providers, err := a.service.providerRepo.List(ctx, "push", notification.TenantID)
	if err != nil {
		a.service.logger.Warn(logging.General, logging.Select,
			"Failed to list push providers from table, falling back to settings",
			map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
	} else if len(providers) > 0 {
		configs := make([]*push.ProviderConfig, 0, len(providers))
		for _, p := range providers {
			if !p.IsEnabled || p.Status == models.ProviderStatusDisabled {
				continue
			}
			cfg, err := providerConfigToPush(p)
			if err != nil {
				a.service.logger.Warn(logging.General, logging.Api,
					"Skipping push provider with invalid config",
					map[logging.ExtraKey]interface{}{
						"provider": p.Name,
						"error":    err.Error(),
					})
				continue
			}
			configs = append(configs, cfg)
		}
		if len(configs) > 0 {
			return configs, nil
		}
	}

	// Legacy fallback: single provider from settings table
	setting, err := a.service.settingRepo.GetByKey(ctx, SettingKeyPushProviders)
	if err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to load push provider config",
			map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
		return nil, fmt.Errorf("push provider not configured: %w", err)
	}

	config, err := push.ParseProviderConfig(setting.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid push provider config: %w", err)
	}

	return []*push.ProviderConfig{config}, nil
}

// GetSMSConfig retrieves SMS config from database
func (s *NotificationService) GetSMSConfig(ctx context.Context) (map[string]string, error) {
	setting, err := s.settingRepo.GetByKey(ctx, models.SettingKeySMSProviders)
	if err != nil {
		return nil, fmt.Errorf("SMS provider config not found in database")
	}

	// Parse the JSON config value to extract the provider name
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(setting.Value), &config); err != nil {
		return nil, fmt.Errorf("invalid SMS provider config JSON: %w", err)
	}

	providerName := ""
	if p, ok := config["provider"].(string); ok {
		providerName = p
	}

	return map[string]string{
		"provider": providerName,
	}, nil
}

// GetEmailConfig retrieves email config from database
func (s *NotificationService) GetEmailConfig(ctx context.Context) (map[string]string, error) {
	setting, err := s.settingRepo.GetByKey(ctx, models.SettingKeyEmailProviders)
	if err != nil {
		return nil, fmt.Errorf("email provider config not found in database")
	}

	// Parse the JSON config value to extract the provider name
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(setting.Value), &config); err != nil {
		return nil, fmt.Errorf("invalid email provider config JSON: %w", err)
	}

	providerName := ""
	if p, ok := config["provider"].(string); ok {
		providerName = p
	}

	return map[string]string{
		"provider": providerName,
	}, nil
}
