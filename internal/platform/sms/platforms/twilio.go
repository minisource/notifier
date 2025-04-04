package providers

import (
	"fmt"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioClient struct {
	template string
	core     *twilio.RestClient
}

func GetTwilioClient(accessId string, accessKey string, template string) (*TwilioClient, error) {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accessId,
		Password: accessKey,
	})

	twilioClient := &TwilioClient{
		core:     client,
		template: template,
	}

	return twilioClient, nil
}

// SendMessage targetPhoneNumber[0] is the sender's number, so targetPhoneNumber should have at least two parameters
func (c *TwilioClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	code, ok := param["code"]
	if !ok {
		return fmt.Errorf("missing parameter: code")
	}

	bodyContent := fmt.Sprintf(c.template, code)

	if len(targetPhoneNumber) < 2 {
		return fmt.Errorf("bad parameter: targetPhoneNumber")
	}

	params := &openapi.CreateMessageParams{}
	params.SetFrom(targetPhoneNumber[0])
	params.SetBody(bodyContent)

	for i := 1; i < len(targetPhoneNumber); i++ {
		params.SetTo(targetPhoneNumber[i])
		_, err := c.core.Api.CreateMessage(params)
		if err != nil {
			return err
		}
	}

	return nil
}
