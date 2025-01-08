package providers

import (
	"github.com/kavenegar/kavenegar-go"
)

func Kavenegar(apiKey, to, body, template string) error {
	api := kavenegar.New(apiKey)
	params := &kavenegar.VerifyLookupParam{}
	if _, err := api.Verify.Lookup(to, template, body, params); err != nil {
		return err;
	}
	
	return nil
}
