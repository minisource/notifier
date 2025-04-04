package providers

import (
	"fmt"

	"github.com/kavenegar/kavenegar-go"
)

type KavenegarClient struct {
	apiKey string
	template  string
	core *kavenegar.Kavenegar
}

func GetKavenegarClient(accessKey string, template string) (*KavenegarClient, error) {
	client := kavenegar.New(accessKey)

	kavenegarClient := &KavenegarClient{
		core:     client,
		template: template,
	}

	return kavenegarClient, nil
}

func (k *KavenegarClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	if len(targetPhoneNumber) == 0 {
		return fmt.Errorf("no target phone number provided")
	}

	params := &kavenegar.VerifyLookupParam{}

	for _, phoneNumber := range targetPhoneNumber {
		message, ok := param["code"]
		if !ok {
			return fmt.Errorf("message parameter is missing")
		}

		if _, err := k.core.Verify.Lookup(phoneNumber, k.template, message, params); err != nil {
			return err
		}
	}

	return nil
}
