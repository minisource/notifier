package providers
type SmsClient interface {
	SendMessage(param map[string]string, targetPhoneNumber ...string) error
}