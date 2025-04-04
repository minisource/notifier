package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ACSClient struct {
	AccessToken string
	Endpoint    string
	Message     string
	Sender      string
}

type reqBody struct {
	From          string         `json:"from"`
	Message       string         `json:"message"`
	SMSRecipients []smsRecipient `json:"smsRecipients"`
}

type smsRecipient struct {
	To string `json:"to"`
}

func GetACSClient(accessToken string, message string, other []string) (*ACSClient, error) {
	if len(other) < 2 {
		return nil, fmt.Errorf("missing parameter: endpoint or sender")
	}

	acsClient := &ACSClient{
		AccessToken: accessToken,
		Endpoint:    other[0],
		Message:     message,
		Sender:      other[1],
	}

	return acsClient, nil
}

func (a *ACSClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	if len(targetPhoneNumber) == 0 {
		return fmt.Errorf("missing parameter: targetPhoneNumber")
	}

	reqBody := &reqBody{
		From:          a.Sender,
		Message:       a.Message,
		SMSRecipients: make([]smsRecipient, 0),
	}
	for _, mobile := range targetPhoneNumber {
		reqBody.SMSRecipients = append(reqBody.SMSRecipients, smsRecipient{To: mobile})
	}

	url := fmt.Sprintf("%s/sms?api-version=2021-03-07", a.Endpoint)

	client := &http.Client{}

	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error creating request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+a.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}

	resp.Body.Close()

	return nil
}
