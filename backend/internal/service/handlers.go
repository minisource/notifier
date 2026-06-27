// Package service provides production-ready notification handling implementations
// SMS, Email, and Push notifications are sent via configurable providers loaded from database

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/minisource/go-common/common"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
	"github.com/minisource/notifier/internal/platform/email"
	"github.com/minisource/notifier/internal/platform/push"
	"github.com/minisource/notifier/internal/platform/sms"
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

	// If no data provided, use body as the code/token
	if len(data) == 0 && notification.Body != "" {
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

	// Load SMS provider config from database
	setting, err := a.service.settingRepo.GetByKey(ctx, models.SettingKeySMSProviders)
	if err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to load SMS provider config from database",
			map[logging.ExtraKey]interface{}{
				"error":      err.Error(),
				"settingKey": models.SettingKeySMSProviders,
			})
		return "", fmt.Errorf("SMS provider not configured: %w", err)
	}

	// Parse provider config
	config, err := sms.ParseProviderConfig(setting.Value)
	if err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to parse SMS provider config",
			map[logging.ExtraKey]interface{}{
				"error":        err.Error(),
				"settingValue": setting.Value,
			})
		return "", fmt.Errorf("invalid SMS provider config: %w", err)
	}

	a.service.logger.Info(logging.General, logging.Api,
		"SMS provider config parsed",
		map[logging.ExtraKey]interface{}{
			"provider": config.Provider,
		})

	// Prepare message parameters
	params := make(map[string]string)

	// Look up SMS template from database if template key is provided
	if templateKey != "" {
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
	} else {
		// No template - use body directly
		params["message"] = notification.Body
		params["body"] = notification.Body
		params["token"] = notification.Body
		params["code"] = notification.Body
	}

	// Create SMS client
	client, err := sms.NewClientFromConfig(config)
	if err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to create SMS client from config",
			map[logging.ExtraKey]interface{}{
				"error":    err.Error(),
				"provider": config.Provider,
			})
		return "", fmt.Errorf("failed to create SMS client: %w", err)
	}

	// For Twilio, we need sender number as first target
	var targets []string
	if config.Provider == "twilio" && config.SenderID != "" {
		targets = append(targets, config.SenderID)
	}
	targets = append(targets, normalizedPhone)

	// Send SMS
	if err := client.SendMessage(params, targets...); err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to send SMS",
			map[logging.ExtraKey]interface{}{
				"notificationID": notification.ID,
				"error":          err.Error(),
			})
		return "", fmt.Errorf("failed to send SMS: %w", err)
	}

	messageID := fmt.Sprintf("sms-%s", notification.ID.String()[:8])
	a.service.logger.Info(logging.General, logging.Api,
		"SMS sent successfully",
		map[logging.ExtraKey]interface{}{
			"notificationID": notification.ID,
			"messageID":      messageID,
			"provider":       config.Provider,
		})

	return messageID, nil
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
	a.service.logger.Info(logging.General, logging.Api,
		"Sending email notification",
		map[logging.ExtraKey]interface{}{
			"notificationID": notification.ID,
			"email":          notification.RecipientEmail,
		})

	// Load email provider config from database
	setting, err := a.service.settingRepo.GetByKey(ctx, models.SettingKeyEmailProviders)
	if err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to load email provider config",
			map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
		return "", fmt.Errorf("email provider not configured: %w", err)
	}

	// Parse provider config
	config, err := email.ParseProviderConfig(setting.Value)
	if err != nil {
		return "", fmt.Errorf("invalid email provider config: %w", err)
	}

	// Create email client
	client, err := email.NewClientFromConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create email client: %w", err)
	}

	// Determine if content is HTML (check for common HTML tags)
	isHTML := len(notification.Body) > 0 && (notification.Body[0] == '<' ||
		len(notification.Body) > 5 && notification.Body[:5] == "<!DOC")

	// Send email
	if err := client.SendEmail(notification.RecipientEmail, notification.Subject, notification.Body, isHTML); err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to send email",
			map[logging.ExtraKey]interface{}{
				"notificationID": notification.ID,
				"error":          err.Error(),
			})
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	messageID := fmt.Sprintf("email-%s", notification.ID.String()[:8])
	a.service.logger.Info(logging.General, logging.Api,
		"Email sent successfully",
		map[logging.ExtraKey]interface{}{
			"notificationID": notification.ID,
			"messageID":      messageID,
		})

	return messageID, nil
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
	a.service.logger.Info(logging.General, logging.Api,
		"Sending push notification",
		map[logging.ExtraKey]interface{}{
			"notificationID": notification.ID,
			"recipientID":    notification.RecipientID,
		})

	// Load push provider config from database
	setting, err := a.service.settingRepo.GetByKey(ctx, SettingKeyPushProviders)
	if err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to load push provider config",
			map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
		return "", fmt.Errorf("push provider not configured: %w", err)
	}

	// Parse provider config
	config, err := push.ParseProviderConfig(setting.Value)
	if err != nil {
		return "", fmt.Errorf("invalid push provider config: %w", err)
	}

	// Create push client
	client, err := push.NewClientFromConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create push client: %w", err)
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
		return "", fmt.Errorf("device token not found for recipient %s", notification.RecipientID)
	}

	// Prepare data payload from metadata
	data := make(map[string]string)
	for k, v := range metadata {
		if str, ok := v.(string); ok && k != "deviceToken" {
			data[k] = str
		}
	}
	data["notificationId"] = notification.ID.String()

	// Send push notification
	if err := client.SendPush(deviceToken, notification.Subject, notification.Body, data); err != nil {
		a.service.logger.Error(logging.General, logging.Api,
			"Failed to send push notification",
			map[logging.ExtraKey]interface{}{
				"notificationID": notification.ID,
				"error":          err.Error(),
			})
		return "", fmt.Errorf("failed to send push notification: %w", err)
	}

	messageID := fmt.Sprintf("push-%s", notification.ID.String()[:8])
	a.service.logger.Info(logging.General, logging.Api,
		"Push notification sent successfully",
		map[logging.ExtraKey]interface{}{
			"notificationID": notification.ID,
			"messageID":      messageID,
		})

	return messageID, nil
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
