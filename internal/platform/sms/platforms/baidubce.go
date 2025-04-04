package providers

import (
	"fmt"
	"strings"

	"github.com/baidubce/bce-sdk-go/services/sms"
	"github.com/baidubce/bce-sdk-go/services/sms/api"
)

type BaiduClient struct {
	sign     string
	template string
	core     *sms.Client
}

func GetBceClient(accessId, accessKey, sign, template string, endpoint []string) (*BaiduClient, error) {
	if len(endpoint) == 0 {
		return nil, fmt.Errorf("missing parameter: endpoint")
	}

	client, err := sms.NewClient(accessId, accessKey, endpoint[0])
	if err != nil {
		return nil, err
	}

	bceClient := &BaiduClient{
		sign:     sign,
		template: template,
		core:     client,
	}

	return bceClient, nil
}

func (c *BaiduClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	code, ok := param["code"]
	if !ok {
		return fmt.Errorf("missing parameter: code")
	}

	if len(targetPhoneNumber) == 0 {
		return fmt.Errorf("missing parameter: targetPhoneNumber")
	}

	contentMap := make(map[string]interface{})
	contentMap["code"] = code

	sendSmsArgs := &api.SendSmsArgs{
		Mobile:      strings.Join(targetPhoneNumber, ","),
		SignatureId: c.sign,
		Template:    c.template,
		ContentVar:  contentMap,
	}

	_, err := c.core.SendSms(sendSmsArgs)
	if err != nil {
		return err
	}

	return nil
}
