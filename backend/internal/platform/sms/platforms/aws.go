package providers

// import (
// 	"fmt"

// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/aws/credentials"
// 	"github.com/aws/aws-sdk-go/aws/session"
// 	"github.com/aws/aws-sdk-go/service/sns"
// 	"github.com/aws/aws-sdk-go/service/sns/snsiface"
// )

// type AmazonSNSClient struct {
// 	svc      snsiface.SNSAPI
// 	template string
// }

// func GetAmazonSNSClient(accessKeyID string, secretAccessKey string, template string, region []string) (*AmazonSNSClient, error) {
// 	if len(region) == 0 {
// 		return nil, fmt.Errorf("missing parameter: region")
// 	}

// 	sess, err := session.NewSession(&aws.Config{
// 		Region:      aws.String(region[0]),
// 		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	svc := sns.New(sess)

// 	snsClient := &AmazonSNSClient{
// 		svc:      svc,
// 		template: template,
// 	}

// 	return snsClient, nil
// }

// func (a *AmazonSNSClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
// 	code, ok := param["code"]
// 	if !ok {
// 		return fmt.Errorf("missing parameter: code")
// 	}

// 	bodyContent := fmt.Sprintf(a.template, code)

// 	if len(targetPhoneNumber) == 0 {
// 		return fmt.Errorf("missing parameter: targetPhoneNumber")
// 	}

// 	messageAttributes := make(map[string]*sns.MessageAttributeValue)
// 	for k, v := range param {
// 		messageAttributes[k] = &sns.MessageAttributeValue{
// 			DataType:    aws.String("String"),
// 			StringValue: aws.String(v),
// 		}
// 	}

// 	for i := 0; i < len(targetPhoneNumber); i++ {
// 		_, err := a.svc.Publish(&sns.PublishInput{
// 			Message:           &bodyContent,
// 			PhoneNumber:       &targetPhoneNumber[i],
// 			MessageAttributes: messageAttributes,
// 		})
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
