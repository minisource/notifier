package push

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PushClient interface for sending push notifications
type PushClient interface {
	SendPush(deviceToken, title, body string, data map[string]string) error
}

// ProviderConfig holds push notification provider configuration
type ProviderConfig struct {
	Provider       string `json:"provider"`       // fcm, apns, onesignal
	ServerKey      string `json:"serverKey"`      // FCM legacy server key
	ProjectID      string `json:"projectId"`      // FCM project ID
	PrivateKeyJSON string `json:"privateKeyJson"` // FCM service account JSON
	APIKey         string `json:"apiKey"`         // OneSignal API key
	AppID          string `json:"appId"`          // OneSignal App ID
	TeamID         string `json:"teamId"`         // APNs team ID
	KeyID          string `json:"keyId"`          // APNs key ID
	BundleID       string `json:"bundleId"`       // APNs bundle ID
	PrivateKey     string `json:"privateKey"`     // APNs private key
	Production     bool   `json:"production"`     // APNs environment
}

// ParseProviderConfig parses JSON config string into ProviderConfig
func ParseProviderConfig(configJSON string) (*ProviderConfig, error) {
	var config ProviderConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to parse push provider config: %w", err)
	}
	return &config, nil
}

// NewClientFromConfig creates a push client based on the provider configuration
func NewClientFromConfig(config *ProviderConfig) (PushClient, error) {
	switch config.Provider {
	case "fcm":
		return NewFCMClient(config)
	case "mock":
		return &MockPushClient{}, nil
	default:
		return nil, fmt.Errorf("unsupported push provider: %s", config.Provider)
	}
}

// FCMClient implements PushClient for Firebase Cloud Messaging (Legacy HTTP API)
type FCMClient struct {
	serverKey string
	client    *http.Client
}

// NewFCMClient creates a new FCM push client using legacy server key
func NewFCMClient(config *ProviderConfig) (*FCMClient, error) {
	if config.ServerKey == "" {
		return nil, fmt.Errorf("FCM server key is required")
	}
	return &FCMClient{
		serverKey: config.ServerKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// FCMMessage represents the FCM message payload
type FCMMessage struct {
	To           string            `json:"to"`
	Notification *FCMNotification  `json:"notification,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
}

// FCMNotification represents the notification part of FCM message
type FCMNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// FCMResponse represents the FCM API response
type FCMResponse struct {
	MessageID int64       `json:"message_id"`
	Success   int         `json:"success"`
	Failure   int         `json:"failure"`
	Results   []FCMResult `json:"results"`
}

// FCMResult represents individual result in FCM response
type FCMResult struct {
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// SendPush sends a push notification via FCM
func (c *FCMClient) SendPush(deviceToken, title, body string, data map[string]string) error {
	msg := FCMMessage{
		To: deviceToken,
		Notification: &FCMNotification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal FCM message: %w", err)
	}

	req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "key="+c.serverKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send FCM request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read FCM response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FCM returned error: %s", string(respBody))
	}

	var fcmResp FCMResponse
	if err := json.Unmarshal(respBody, &fcmResp); err != nil {
		return fmt.Errorf("failed to parse FCM response: %w", err)
	}

	if fcmResp.Failure > 0 && len(fcmResp.Results) > 0 {
		return fmt.Errorf("FCM send failed: %s", fcmResp.Results[0].Error)
	}

	return nil
}

// MockPushClient is a mock implementation for testing
type MockPushClient struct{}

// SendPush mock implementation
func (c *MockPushClient) SendPush(deviceToken, title, body string, data map[string]string) error {
	// Mock implementation - just log
	return nil
}
