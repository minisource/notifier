package notification

import (
	"github.com/minisource/common_go/logging"
	"github.com/minisource/notifier/api/v1/dto"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/platform/sms"
)

type SMSService struct {
	logger logging.Logger
	cfg    *config.Config
}

func NewSMSService(cfg *config.Config) *SMSService {
	logger := logging.NewLogger(&cfg.Logger)
	return &SMSService{
		logger: logger,
		cfg:    cfg,
	}
}

func (s *SMSService) SendNotification(req dto.SMSRequest) error {
	if s.cfg.SMS.NotEnabled {
		return nil
	}

	if req.To == "" {
		req.To = req.PhoneNumber
	}
	return sms.SendSMS(&s.cfg.SMS, s.logger, req.To, req.Body, req.Template)
}
