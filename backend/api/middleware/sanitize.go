package middleware

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

// SensitiveKeys lists JSON keys that should be redacted/masked in responses and logs.
var SensitiveKeys = map[string]bool{
	"password":       true,
	"token":          true,
	"secret":         true,
	"apiKey":         true,
	"apikey":         true,
	"api_key":        true,
	"authorization":  true,
	"auth":           true,
	"otp":            true,
	"code":           true,
	"refreshToken":   true,
	"refresh_token":  true,
	"accessToken":    true,
	"access_token":   true,
	"privateKey":     true,
	"private_key":    true,
	"webhookSecret":  true,
	"webhook_secret": true,
	"jwt":            true,
	"jwt_secret":     true,
	"clientSecret":   true,
	"client_secret":  true,
	"sessionToken":   true,
	"session_token":  true,
}

// MaskEmail masks an email address for PII protection.
// E.g., "user@example.com" → "u***r@example.com"
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***"
	}
	local := parts[0]
	domain := parts[1]
	if len(local) <= 2 {
		return local[:1] + "***@" + domain
	}
	return local[:1] + "***" + local[len(local)-1:] + "@" + domain
}

// MaskPhone masks a phone number for PII protection.
// E.g., "+989121234567" → "+98********67"
func MaskPhone(phone string) string {
	if phone == "" || len(phone) < 4 {
		return phone
	}
	masked := phone[:2] + strings.Repeat("*", len(phone)-4) + phone[len(phone)-2:]
	return masked
}

// MaskRecipient returns a masked recipient string based on type heuristic.
func MaskRecipient(email, phone, userID string) string {
	if email != "" {
		return MaskEmail(email)
	}
	if phone != "" {
		return MaskPhone(phone)
	}
	if userID != "" {
		if len(userID) > 8 {
			return userID[:8] + "***"
		}
		return userID
	}
	return ""
}

// SanitizeProviderResponse redacts sensitive fields from provider response JSON.
func SanitizeProviderResponse(response string) string {
	if response == "" || response == "{}" {
		return ""
	}
	// Try to parse as JSON and redact sensitive keys
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		// Not valid JSON — truncate to safe length
		if len(response) > 500 {
			return response[:500] + "... [truncated]"
		}
		return response
	}

	// Redact sensitive keys
	for key := range data {
		lowerKey := strings.ToLower(key)
		if SensitiveKeys[lowerKey] {
			data[key] = "[REDACTED]"
		}
	}

	sanitized, _ := json.Marshal(data)
	return string(sanitized)
}

// SanitizeErrorMessage filters sensitive information from error messages.
func SanitizeErrorMessage(msg string) string {
	if msg == "" {
		return ""
	}
	// Redact patterns that look like secrets
	lower := strings.ToLower(msg)
	for key := range SensitiveKeys {
		if strings.Contains(lower, key) {
			// Replace the entire message with a safe summary
			return "[Error message redacted — may contain sensitive information]"
		}
	}
	// Truncate long messages
	if len(msg) > 500 {
		return msg[:500] + "... [truncated]"
	}
	return msg
}

// SanitizeMetadata redacts sensitive keys from a metadata JSON string.
func SanitizeMetadata(metadata string) string {
	if metadata == "" || metadata == "{}" {
		return "{}"
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(metadata), &data); err != nil {
		return "{}"
	}

	for key := range data {
		lowerKey := strings.ToLower(key)
		if SensitiveKeys[lowerKey] {
			data[key] = "[REDACTED]"
		}
	}

	sanitized, _ := json.Marshal(data)
	return string(sanitized)
}

// SanitizeAuditPayload redacts sensitive fields from an audit log payload.
func SanitizeAuditPayload(payload map[string]interface{}) map[string]interface{} {
	if payload == nil {
		return nil
	}
	result := make(map[string]interface{}, len(payload))
	for key, val := range payload {
		lowerKey := strings.ToLower(key)
		if SensitiveKeys[lowerKey] {
			result[key] = "[REDACTED]"
		} else if str, ok := val.(string); ok && len(str) > 500 {
			result[key] = str[:500] + "..."
		} else {
			result[key] = val
		}
	}
	return result
}

// SanitizeNotificationItem sanitizes a notification list item for safe external display.
func SanitizeNotificationItem(id uuid.UUID, userID uuid.UUID, nType string, status string, subject string, createdAt interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":        id.String(),
		"userId":    userID.String(),
		"type":      nType,
		"status":    status,
		"subject":   subject,
		"createdAt": createdAt,
	}
}
