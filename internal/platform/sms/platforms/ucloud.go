package providers

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/usms"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/config"
)

type UcloudClient struct {
	core       *usms.USMSClient
	ProjectId  string
	PrivateKey string
	PublicKey  string
	Sign       string
	Template   string
}

func GetUcloudClient(publicKey string, privateKey string, sign string, template string, projectId []string) (*UcloudClient, error) {
	if len(projectId) == 0 {
		return nil, fmt.Errorf("missing parameter: projectId")
	}

	cfg := config.NewConfig()
	cfg.ProjectId = projectId[0]
	credential := auth.NewCredential()
	credential.PublicKey = publicKey
	credential.PrivateKey = privateKey

	client := usms.NewClient(&cfg, &credential)

	ucloudClient := &UcloudClient{
		core:       client,
		ProjectId:  projectId[0],
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Sign:       sign,
		Template:   template,
	}

	return ucloudClient, nil
}

func (c *UcloudClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	code, ok := param["code"]
	if !ok {
		return fmt.Errorf("missing parameter: code")
	}

	if len(targetPhoneNumber) == 0 {
		return fmt.Errorf("missing parameter: targetPhoneNumber")
	}

	req := c.core.NewSendUSMSMessageRequest()
	req.SigContent = ucloud.String(c.Sign)
	req.TemplateId = ucloud.String(c.Template)
	req.PhoneNumbers = targetPhoneNumber
	req.TemplateParams = []string{code}
	response, err := c.core.SendUSMSMessage(req)
	if err != nil {
		return err
	}
	if response.RetCode != 0 {
		return fmt.Errorf(response.Message)
	}
	return nil
}
