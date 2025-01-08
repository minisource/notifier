package providers

type SendSMS interface {
	Send(apiKey, to, body, template string) error
}
