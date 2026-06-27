package providers

import (
	"errors"
	"fmt"
	"strings"

	unisms "github.com/apistd/uni-go-sdk/sms"
)

type UnismsClient struct {
	core     *unisms.UniSMSClient
	sign     string
	template string
}

func GetUnismsClient(accessId string, accessKey string, signature string, templateId string) (*UnismsClient, error) {
	client := unisms.NewClient(accessId, accessKey)

	// Check the correctness of the accessId and accessKey
	msg := unisms.BuildMessage()
	msg.SetTo("test")
	msg.SetTemplateId("pub_verif_register") // free template
	_, err := client.Send(msg)
	if strings.Contains(err.Error(), "[104111] InvalidAccessKeyId") {
		return nil, err
	}

	unismsClient := &UnismsClient{
		core:     client,
		sign:     signature,
		template: templateId,
	}

	return unismsClient, nil
}

func (c *UnismsClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	if len(targetPhoneNumber) == 0 {
		return fmt.Errorf("missing parameter: targetPhoneNumber")
	}

	msg := unisms.BuildMessage()
	msg.SetTo(targetPhoneNumber...)
	msg.SetSignature(c.sign)
	msg.SetTemplateId(c.template)

	resp, err := c.core.Send(msg)
	if err != nil {
		return err
	}

	if resp.Code != "0" {
		return errors.New(resp.Message)
	}

	return nil
}
