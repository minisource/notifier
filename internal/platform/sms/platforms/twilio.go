package providers

import "fmt"

// Twilio represents the Twilio SMS provider
type TwilioServcie struct {
	AccountSID string
	AuthToken  string
}

// TODO: SendSMS sends an SMS via the Twilio provider
func (t *TwilioServcie) SendSMS(to, message string) error {
	// Simulating the sending of SMS via Twilio
	fmt.Printf("Sending SMS to %s via Twilio: %s\n", to, message)
	return nil
}
