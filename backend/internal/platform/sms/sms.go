package sms

import (
	"encoding/json"
	"fmt"

	providers "github.com/minisource/notifier/internal/platform/sms/platforms"
)

// ProviderConfig holds SMS provider configuration loaded from database
type ProviderConfig struct {
	Provider  string   `json:"provider"`
	APIKey    string   `json:"apiKey"`
	AccessID  string   `json:"accessId"`
	AccessKey string   `json:"accessKey"`
	Sign      string   `json:"sign"`
	Template  string   `json:"template"`
	SenderID  string   `json:"senderId"`
	Extra     []string `json:"extra"` // Extra params like region, apiAddress, etc.
}

// ParseProviderConfig parses JSON config string into ProviderConfig
func ParseProviderConfig(configJSON string) (*ProviderConfig, error) {
	var config ProviderConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to parse provider config: %w", err)
	}
	return &config, nil
}

// NewClientFromConfig creates an SMS client based on the provider configuration
// Only providers with active implementations are supported
func NewClientFromConfig(config *ProviderConfig) (providers.SmsClient, error) {
	switch config.Provider {
	case "kavenegar":
		return providers.GetKavenegarClient(config.APIKey, config.Template)
	case "twilio":
		return providers.GetTwilioClient(config.AccessID, config.AccessKey, config.Template)
	case "tencent":
		// Tencent expects: accessId, accessKey, sign, templateId, appId[]
		return providers.GetTencentClient(config.AccessID, config.AccessKey, config.Sign, config.Template, config.Extra)
	case "huawei":
		// Huawei expects: accessId, accessKey, sign, template, other[] (apiAddress, sender)
		return providers.GetHuaweiClient(config.AccessID, config.AccessKey, config.Sign, config.Template, config.Extra)
	case "infobip":
		// Infobip expects: sender, apiKey, template, baseUrl[]
		return providers.GetInfobipClient(config.SenderID, config.APIKey, config.Template, config.Extra)
	case "msg91":
		// msg91 expects: senderId, authKey, templateId
		return providers.GetMsg91Client(config.SenderID, config.APIKey, config.Template)
	case "netgsm":
		return providers.GetNetgsmClient(config.AccessID, config.AccessKey, config.SenderID, config.Template)
	case "oson":
		return providers.GetOsonClient(config.AccessID, config.AccessKey, config.SenderID, config.Template)
	case "smsbao":
		// smsbao expects: username, apikey, sign, template, other[]
		return providers.GetSmsbaoClient(config.AccessID, config.AccessKey, config.Sign, config.Template, config.Extra)
	case "submail":
		return providers.GetSubmailClient(config.AccessID, config.AccessKey, config.Template)
	case "mock":
		return providers.NewMocker(config.AccessID, config.AccessKey, config.Sign, config.Template, config.Extra)
	default:
		return nil, fmt.Errorf("unsupported or inactive SMS provider: %s (supported: kavenegar, twilio, tencent, huawei, infobip, msg91, netgsm, oson, smsbao, submail, mock)", config.Provider)
	}
}
