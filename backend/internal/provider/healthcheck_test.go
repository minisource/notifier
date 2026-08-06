package provider

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/minisource/notifier/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMergeProviderConfigJSON_InjectTypeFromColumn guards the fix for the
// provider test endpoint: providers created through the admin UI store only
// channel fields (apiKey, template, ...) in Config/SecretConfig — the
// "provider" key is omitted. The client factories switch on "provider", so
// MergeProviderConfigJSON must inject it from the row's Type column.
func TestMergeProviderConfigJSON_InjectTypeFromColumn(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		secrets  string
		typ      string
		expected interface{}
	}{
		{
			name:     "injects provider from Type when config omits it",
			config:   `{"template":"verify"}`,
			secrets:  `{"apiKey":"secret-key"}`,
			typ:      "kavenegar",
			expected: "kavenegar",
		},
		{
			name:     "keeps explicit provider key when present",
			config:   `{"provider":"twilio","template":"x"}`,
			typ:      "kavenegar",
			expected: "twilio",
		},
		{
			name:     "no provider injected when Type is empty",
			config:   `{"template":"verify"}`,
			typ:      "",
			expected: nil,
		},
		{
			name:     "works with empty config entirely",
			typ:      "twilio",
			expected: "twilio",
		},
		{
			name:     "lowercases mixed-case Type so factory switches match",
			config:   `{"template":"verify"}`,
			typ:      "Kavenegar",
			expected: "kavenegar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &models.Provider{
				ID:           uuid.New(),
				Name:         "test-provider",
				Channel:      "sms",
				Type:         tt.typ,
				Config:       tt.config,
				SecretConfig: tt.secrets,
			}
			raw, err := MergeProviderConfigJSON(p)
			require.NoError(t, err)

			var merged map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(raw), &merged))
			got, present := merged["provider"]
			if tt.expected == nil {
				assert.False(t, present, "provider key should not be injected when Type is empty")
				return
			}
			assert.True(t, present)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestMergeProviderConfigJSON_SecretsWin ensures SecretConfig still overrides
// Config on key conflicts after the provider injection change.
func TestMergeProviderConfigJSON_SecretsWin(t *testing.T) {
	p := &models.Provider{
		ID:           uuid.New(),
		Name:         "test",
		Channel:      "sms",
		Type:         "kavenegar",
		Config:       `{"apiKey":"public-placeholder","template":"verify"}`,
		SecretConfig: `{"apiKey":"real-secret"}`,
	}
	raw, err := MergeProviderConfigJSON(p)
	require.NoError(t, err)

	var merged map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(raw), &merged))
	assert.Equal(t, "real-secret", merged["apiKey"])
	assert.Equal(t, "verify", merged["template"])
	assert.Equal(t, "kavenegar", merged["provider"])
}
