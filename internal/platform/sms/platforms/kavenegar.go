package providers

import (
	"github.com/kavenegar/kavenegar-go"
)

type KavenegarService struct {
	ApiKey string
	Template  string
}


func (k *KavenegarService) SendSMS(to, message string) error {
	api := kavenegar.New(k.ApiKey)
	params := &kavenegar.VerifyLookupParam{}
	if _, err := api.Verify.Lookup(to, k.Template, message, params); err != nil {
		return err;
	}
	
	return nil
}
