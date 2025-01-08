package sms

import (
	"github.com/minisource/common_go/logging"
	"github.com/minisource/notifire/config"
	providers "github.com/minisource/notifire/internal/platform/sms/platforms"
)

// متد ارسال پیامک
func SendSMS(cfg *config.SMSConfig, logger logging.Logger, to, body, template string) error {
	switch cfg.Provider {
	default:
		err := providers.Kavenegar(cfg.ApiKey, to, body, template)
		if err != nil {
			return err
		}
	}

	return nil
}
