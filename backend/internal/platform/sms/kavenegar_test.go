package sms

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKavenegar_SendWithTemplateLookup(t *testing.T) {
	apiKey := os.Getenv("KAVENEGAR_API_KEY")
	if apiKey == "" {
		t.Skip("KAVENEGAR_API_KEY not set")
	}

	testPhone := os.Getenv("TEST_PHONE")
	if testPhone == "" {
		testPhone = "09011793041" // Default test phone
	}

	// Create provider config
	config := &ProviderConfig{
		Provider: "kavenegar",
		APIKey:   apiKey,
		Template: "verify",
	}

	// Create SMS client
	client, err := NewClientFromConfig(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test sending with template lookup
	params := map[string]string{
		"token": "123456",
	}

	err = client.SendMessage(params, testPhone)
	assert.NoError(t, err, "Should send SMS successfully via Kavenegar lookup API")

	t.Logf("✅ SMS sent successfully to %s with token 123456", testPhone)
}

func TestKavenegar_SendWithMultipleTokens(t *testing.T) {
	apiKey := os.Getenv("KAVENEGAR_API_KEY")
	if apiKey == "" {
		t.Skip("KAVENEGAR_API_KEY not set")
	}

	testPhone := os.Getenv("TEST_PHONE")
	if testPhone == "" {
		testPhone = "09011793041"
	}

	config := &ProviderConfig{
		Provider: "kavenegar",
		APIKey:   apiKey,
		Template: "verify",
	}

	client, err := NewClientFromConfig(config)
	require.NoError(t, err)

	// Test with multiple tokens
	params := map[string]string{
		"token":  "123456",
		"token2": "TestValue",
	}

	err = client.SendMessage(params, testPhone)
	assert.NoError(t, err, "Should send SMS with multiple tokens")

	t.Logf("✅ SMS sent with multiple tokens: token=123456, token2=TestValue")
}

func TestKavenegar_InvalidPhone(t *testing.T) {
	apiKey := os.Getenv("KAVENEGAR_API_KEY")
	if apiKey == "" {
		t.Skip("KAVENEGAR_API_KEY not set")
	}

	config := &ProviderConfig{
		Provider: "kavenegar",
		APIKey:   apiKey,
		Template: "verify",
	}

	client, err := NewClientFromConfig(config)
	require.NoError(t, err)

	params := map[string]string{
		"token": "123456",
	}

	err = client.SendMessage(params, "invalid_phone")
	assert.Error(t, err, "Should fail with invalid phone number")
}

func TestKavenegar_MissingRequiredParams(t *testing.T) {
	apiKey := os.Getenv("KAVENEGAR_API_KEY")
	if apiKey == "" {
		t.Skip("KAVENEGAR_API_KEY not set")
	}

	config := &ProviderConfig{
		Provider: "kavenegar",
		APIKey:   apiKey,
		Template: "verify",
	}

	client, err := NewClientFromConfig(config)
	require.NoError(t, err)

	// Missing token parameter
	params := map[string]string{}

	err = client.SendMessage(params, "09011793041")
	assert.Error(t, err, "Should fail without token parameter")
	assert.Contains(t, err.Error(), "token")
}

func TestKavenegar_ConfigValidation(t *testing.T) {
	// Test with invalid provider
	config := &ProviderConfig{
		Provider: "invalid",
		APIKey:   "test-key",
		Template: "verify",
	}

	client, err := NewClientFromConfig(config)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestKavenegar_BothPhoneFormats(t *testing.T) {
	apiKey := os.Getenv("KAVENEGAR_API_KEY")
	if apiKey == "" {
		t.Skip("KAVENEGAR_API_KEY not set")
	}

	config := &ProviderConfig{
		Provider: "kavenegar",
		APIKey:   apiKey,
		Template: "verify",
	}

	client, err := NewClientFromConfig(config)
	require.NoError(t, err)

	params := map[string]string{
		"token": "999888",
	}

	// Test with 09011793041 format
	t.Run("WithoutCountryCode", func(t *testing.T) {
		err = client.SendMessage(params, "09011793041")
		assert.NoError(t, err, "Should accept 09011793041 format")
		t.Logf("✅ Sent to 09011793041")
	})

	// Test with +989011793041 format
	t.Run("WithCountryCode", func(t *testing.T) {
		err = client.SendMessage(params, "+989011793041")
		assert.NoError(t, err, "Should accept +989011793041 format")
		t.Logf("✅ Sent to +989011793041")
	})
}
