package sms

import (
	"fmt"

	"github.com/minisource/common_go/common"
	"github.com/minisource/common_go/logging"
	"github.com/minisource/notifier/config"
	providers "github.com/minisource/notifier/internal/platform/sms/platforms"
)

const (
	Kavenegar    = "Kavenegar SMS"
	Twilio       = "Twilio SMS"
	AmazonSNS    = "Amazon SNS"
	AzureACS     = "Azure ACS"
	Msg91        = "Msg91 SMS"
	GCCPAY       = "GCCPAY SMS"
	Infobip      = "Infobip SMS"
	SUBMAIL      = "SUBMAIL SMS"
	SmsBao       = "SmsBao SMS"
	Aliyun       = "Aliyun SMS"
	TencentCloud = "Tencent Cloud SMS"
	BaiduCloud   = "Baidu Cloud SMS"
	VolcEngine   = "Volc Engine SMS"
	HuaweiCloud  = "Huawei Cloud SMS"
	UCloud       = "UCloud SMS"
	Huyi         = "Huyi SMS"
	MockSms      = "Mock SMS"
	Netgsm       = "Netgsm SMS"
	OsonSms      = "OSON SMS"
	UniSms       = "Uni SMS"
)

func SendSMS(cfg *config.SMSConfig, logger logging.Logger, to, body, template string) error {
	var client providers.SmsClient

	configs, err := cfg.GetProviderConfig(cfg.DefualtProvider)
	if err != nil {
		logger.Errorf("failed to send SMS: %v", err)
		return err
	}
	if template == "" {
		template = configs.Template
	}

	if common.IsIranianNumber(to) {
		client, err = NewSmsClient("kavenegar", configs.AccessId, configs.AccessKey, configs.Sign, template)
	} else {
		client, err = NewSmsClient(cfg.DefualtProvider, configs.AccessId, configs.AccessKey, configs.Sign, template)
	}

	if err != nil {
		return err
	}
	params := map[string]string{}
	params["code"] = body
	phoneNumer := to
	err = client.SendMessage(params, phoneNumer)
	if err != nil {
		return err
	}

	return nil
}

func NewSmsClient(provider string, accessId string, accessKey string, sign string, template string, other ...string) (providers.SmsClient, error) {
	switch provider {
	case Kavenegar:
		return providers.GetKavenegarClient(accessKey, template)
	case Twilio:
		return providers.GetTwilioClient(accessId, accessKey, template)
	// case AmazonSNS:
	// 	return providers.GetAmazonSNSClient(accessId, accessKey, template, other)
	case AzureACS:
		return providers.GetACSClient(accessKey, template, other)
	case Msg91:
		return providers.GetMsg91Client(accessId, accessKey, template)
	case GCCPAY:
		return providers.GetGCCPAYClient(accessId, accessKey, template)
	case Infobip:
		return providers.GetInfobipClient(accessId, accessKey, template, other)
	case SUBMAIL:
		return providers.GetSubmailClient(accessId, accessKey, template)
	case SmsBao:
		return providers.GetSmsbaoClient(accessId, accessKey, sign, template, other)
	// case Aliyun:
	// 	return providers.GetAliyunClient(accessId, accessKey, sign, template)
	case TencentCloud:
		return providers.GetTencentClient(accessId, accessKey, sign, template, other)
	case BaiduCloud:
		return providers.GetBceClient(accessId, accessKey, sign, template, other)
	case VolcEngine:
		return providers.GetVolcClient(accessId, accessKey, sign, template, other)
	case HuaweiCloud:
		return providers.GetHuaweiClient(accessId, accessKey, sign, template, other)
	case UCloud:
		return providers.GetUcloudClient(accessId, accessKey, sign, template, other)
	case Huyi:
		return providers.GetHuyiClient(accessId, accessKey, template)
	case Netgsm:
		return providers.GetNetgsmClient(accessId, accessKey, sign, template)
	case MockSms:
		return providers.NewMocker(accessId, accessKey, sign, template, other)
	case OsonSms:
		return providers.GetOsonClient(accessId, accessKey, sign, template)
	case UniSms:
		return providers.GetUnismsClient(accessId, accessKey, sign, template)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
