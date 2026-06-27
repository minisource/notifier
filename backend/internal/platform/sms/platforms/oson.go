package providers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type OsonClient struct {
	Endpoint         string
	SenderID         string
	SecretAccessHash string
	Sign             string
	Message          string
}

type OsonResponse struct {
	Status        string    // ok
	Timestamp     time.Time // 2017-07-07 16:58:12
	TxnId         string    // f89xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxe0b
	MsgId         uint      // 40127
	SmscMsgId     string    // 45f22479
	SmscMsgStatus string    // success
	SmscMsgParts  string    // 1
}

func GetOsonClient(senderId, secretAccessHash, sign, message string) (*OsonClient, error) {
	return &OsonClient{
		Endpoint:         "https://api.osonsms.com/sendsms_v1.php",
		SenderID:         senderId,
		SecretAccessHash: secretAccessHash,
		Sign:             sign,
		Message:          message,
	}, nil
}

func (c *OsonClient) SendMessage(param map[string]string, targetPhoneNumber ...string) (err error) {
	// Init http client for make request to sms center. Set a timeout of 25+
	// seconds to ensure that the response from the SMS center has been
	// processed.
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	if c.Message == "" {
		c.Message = fmt.Sprintf("Hello. Your authorization code: %s", param["code"])
	} else {
		c.Message += param["code"]
	}

	txnID := uuid.New()
	buildStrHash := strings.Join([]string{txnID.String(), c.SenderID, c.Sign, targetPhoneNumber[0], c.SecretAccessHash}, ";")

	hash := sha256.New()
	hash.Write([]byte(buildStrHash))
	bs := hash.Sum(nil)
	strHash := fmt.Sprintf("%x", bs)

	urlLink, err := url.Parse(c.Endpoint)
	if err != nil {
		return
	}

	urlParams := url.Values{}
	urlParams.Add("from", c.Sign)
	urlParams.Add("phone_number", targetPhoneNumber[0])
	urlParams.Add("msg", c.Message)
	urlParams.Add("str_hash", strHash)
	urlParams.Add("txn_id", txnID.String())
	urlParams.Add("login", c.SenderID)

	urlLink.RawQuery = urlParams.Encode()

	request, err := http.NewRequest(http.MethodGet, urlLink.String(), nil)
	if err != nil {
		return
	}

	resp, err := client.Do(request)
	if err != nil {
		return
	}

	resultBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result OsonResponse
	if err = json.Unmarshal(resultBytes, &result); err != nil {
		return
	}

	if result.Status != "ok" {
		return fmt.Errorf("sms service returned error status not 200: Status Code: %d Error: %s", resp.StatusCode, string(resultBytes))
	}

	return
}
