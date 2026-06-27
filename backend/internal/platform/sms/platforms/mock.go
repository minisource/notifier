package providers

type Mocker struct{}

var _ SmsClient = &Mocker{}

func NewMocker(accessId, accessKey, sign, templateId string, smsAccount []string) (*Mocker, error) {
	return &Mocker{}, nil
}

func (m *Mocker) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	return nil
}
