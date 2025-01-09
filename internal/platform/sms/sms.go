package sms

import (
	"github.com/minisource/common_go/common"
	"github.com/minisource/common_go/logging"
	"github.com/minisource/notifier/config"
	providers "github.com/minisource/notifier/internal/platform/sms/platforms"
)

func SendSMS(cfg *config.SMSConfig, logger logging.Logger, to, body, template string) error {
	var provider providers.SMSSender

	if cfg.DefualtProvider != "" {
		provider = GetSMSSender(cfg, cfg.DefualtProvider, template)
	} else if common.IsIranianNumber(to) {
		provider = GetSMSSender(cfg, "kavenegar", template)
	} else {
		provider = GetSMSSender(cfg, "twilio", template)
	}

	if err := provider.SendSMS(to, body); err != nil {
		logger.Errorf("failed to send SMS: %v", err)
		return err
	}

	return nil
}

func GetSMSSender(cfg *config.SMSConfig, providerType, template string) providers.SMSSender {
	apiKey, _ := cfg.GetApiKeyByProvider(providerType)

	switch providerType {
	case "kavenegar":
		return &providers.KavenegarService{ApiKey: apiKey, Template: template}
	case "twilio": // TODO
		return &providers.TwilioServcie{AccountSID: "your-account-sid", AuthToken: apiKey}
	default:
		return nil
	}
}
