package providers

import (
	"fmt"

	"github.com/kavenegar/kavenegar-go"
)

type KavenegarClient struct {
	apiKey   string
	template string
	core     *kavenegar.Kavenegar
}

func GetKavenegarClient(accessKey string, template string) (*KavenegarClient, error) {
	client := kavenegar.New(accessKey)

	kavenegarClient := &KavenegarClient{
		core:     client,
		apiKey:   accessKey,
		template: template,
	}

	return kavenegarClient, nil
}

// SendMessage sends SMS via Kavenegar Lookup API (verification/OTP messages)
// Kavenegar uses pre-defined templates in their panel. We only send template name + token values.
//
// Supported params:
//   - template: Template name (optional, overrides default template) e.g., "verify", "orderPlaced"
//   - token/code: Primary token (required) - the main value like OTP code
//   - token2: Secondary token (optional) - e.g., amount, status
//   - token3: Third token (optional)
//   - token10: Token with 5 spaces allowed (optional)
//   - token20: Token with 8 spaces allowed (optional)
//
// Example templates from Kavenegar panel:
//   - "verify": OTP verification, uses %token for code
//   - "orderPlaced": Order placed, uses %token for order number, %token2 for amount
//   - "paymentSuccess": Payment success, uses %token for order, %token2 for amount
func (k *KavenegarClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	if len(targetPhoneNumber) == 0 {
		return fmt.Errorf("no target phone number provided")
	}

	// Get the primary token (code/token)
	token, ok := param["token"]
	if !ok {
		token, ok = param["code"]
		if !ok {
			// Fall back to "message" if neither provided
			token, ok = param["message"]
			if !ok {
				return fmt.Errorf("token, code, or message parameter is required")
			}
		}
	}

	// Determine template to use (param template overrides default)
	template := k.template
	if paramTemplate, ok := param["template"]; ok && paramTemplate != "" {
		template = paramTemplate
	}
	if template == "" {
		return fmt.Errorf("no template specified (set in config or pass as parameter)")
	}

	// Build lookup params with optional tokens
	params := &kavenegar.VerifyLookupParam{
		Tokens: make(map[string]string),
	}
	if token2, ok := param["token2"]; ok && token2 != "" {
		params.Token2 = token2
	}
	if token3, ok := param["token3"]; ok && token3 != "" {
		params.Token3 = token3
	}
	// Token10 and Token20 are sent via the Tokens map
	if token10, ok := param["token10"]; ok && token10 != "" {
		params.Tokens["token10"] = token10
	}
	if token20, ok := param["token20"]; ok && token20 != "" {
		params.Tokens["token20"] = token20
	}

	// Send to all target phone numbers using the lookup API
	// Note: We only send template name + tokens, not full message text
	for _, phoneNumber := range targetPhoneNumber {
		if _, err := k.core.Verify.Lookup(phoneNumber, template, token, params); err != nil {
			return fmt.Errorf("kavenegar lookup failed for %s with template '%s': %w", phoneNumber, template, err)
		}
	}

	return nil
}
