package providers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/volcengine/volc-sdk-golang/service/sms"
)

type VolcClient struct {
	core       *sms.SMS
	sign       string
	template   string
	smsAccount string
}

func GetVolcClient(accessId, accessKey, sign, templateId string, smsAccount []string) (*VolcClient, error) {
	if len(smsAccount) == 0 {
		return nil, fmt.Errorf("missing parameter: smsAccount")
	}

	client := sms.NewInstance()
	client.Client.SetAccessKey(accessId)
	client.Client.SetSecretKey(accessKey)

	volcClient := &VolcClient{
		core:       client,
		sign:       sign,
		template:   templateId,
		smsAccount: smsAccount[0],
	}

	return volcClient, nil
}

func (c *VolcClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	if len(targetPhoneNumber) == 0 {
		return fmt.Errorf("missing parameter: targetPhoneNumber")
	}

	requestParam, err := json.Marshal(param)
	if err != nil {
		return err
	}

	req := &sms.SmsRequest{
		SmsAccount:    c.smsAccount,
		Sign:          c.sign,
		TemplateID:    c.template,
		TemplateParam: string(requestParam),
		PhoneNumbers:  strings.Join(targetPhoneNumber, ","),
	}

	resp, statusCode, err := c.core.Send(req)
	if err != nil {
		return fmt.Errorf("send message failed, error: %q", err.Error())
	}
	if statusCode < 200 || statusCode > 299 {
		return fmt.Errorf("send message failed, statusCode: %d", statusCode)
	}
	if resp.ResponseMetadata.Error != nil {
		return fmt.Errorf("send message failed, code: %q, message: %q", resp.ResponseMetadata.Error.Code, resp.ResponseMetadata.Error.Message)
	}

	return nil
}
