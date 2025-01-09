package providers

import (
	"errors"
)

// Twilio represents the Twilio SMS provider
type TwilioServcie struct {
	AccountSID string
	AuthToken  string
}

// TODO: SendSMS sends an SMS via the Twilio provider
func (t *TwilioServcie) SendSMS(to, message string) error {
	return errors.New("twilio not implemented yet.")
}
