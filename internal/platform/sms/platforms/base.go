package providers

type SMSSender interface {
	SendSMS(to, message string) error
}