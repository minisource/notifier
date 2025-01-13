package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/minisource/common_go/logging"
)

type Config struct {
	Server ServerConfig
	// Postgres PostgresConfig
	Cors     CorsConfig
	Logger   logging.LoggerConfig
	SMS      SMSConfig
	OAUTHURL string `env:"APICLIENTS_OAUTH_URL"`
}

type ServerConfig struct {
	InternalPort string `env:"SERVER_INTERNAL_PORT"`
	ExternalPort string `env:"SERVER_EXTERNAL_PORT"`
	RunMode      string `env:"SERVER_RUN_MODE"`
	Name         string `env:"SERVER_NAME"`
}

// type PostgresConfig struct {
// 	Host            string        `env:"POSTGRES_HOST"`
// 	Port            string        `env:"POSTGRES_PORT"`
// 	User            string        `env:"POSTGRES_USER"`
// 	Password        string        `env:"POSTGRES_PASSWORD"`
// 	DbName          string        `env:"POSTGRES_DBNAME"`
// 	SSLMode         string        `env:"POSTGRES_SSLMODE"`
// 	MaxIdleConns    int           `env:"POSTGRES_MAX_IDLE_CONNS"`
// 	MaxOpenConns    int           `env:"POSTGRES_MAX_OPEN_CONNS"`
// 	ConnMaxLifetime time.Duration `env:"POSTGRES_CONN_MAX_LIFETIME"`
// }

type CorsConfig struct {
	AllowOrigins string `env:"CORS_ALLOW_ORIGINS"`
}

type SMSConfig struct {
	Providers       []SMSProviderConfig
	DefualtProvider string `env:"SMS_Defualt_Provider"`
	NotSendSms      bool   `env:"SMS_NOT_SEND_SMS"`
}

type SMSProviderConfig struct {
	Provider string `env:"SMS_PROVIDER"`
	ApiKey   string `env:"SMS_API_KEY"`
}

// LoadConfig loads configuration from environment variables
func GetConfig() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Error in parse config %v", err)
		panic(err)
	}

	// Handle SMS Providers and API Keys parsing
	smsProviders := os.Getenv("SMS_PROVIDERS")
	if smsProviders != "" {
		smsApiKeys := os.Getenv("SMS_API_KEYS")
		if smsApiKeys != "" {
			providers := strings.Split(smsProviders, ",")
			apiKeys := strings.Split(smsApiKeys, ",")

			if len(providers) != len(apiKeys) {
				log.Fatalf("SMS_PROVIDERS and SMS_API_KEYS must have the same length")
				return nil
			}

			var smsConfigs []SMSProviderConfig
			for i := 0; i < len(providers); i++ {
				smsConfigs = append(smsConfigs, SMSProviderConfig{
					Provider: providers[i],
					ApiKey:   apiKeys[i],
				})
			}
			cfg.SMS.Providers = smsConfigs
		}
	}

	return cfg
}

// GetApiKeyByProvider returns the API key for the given provider
func (cfg *SMSConfig) GetApiKeyByProvider(provider string) (string, error) {
	// Iterate over the SMS configurations
	for _, sms := range cfg.Providers {
		if strings.EqualFold(sms.Provider, provider) {
			return sms.ApiKey, nil
		}
	}

	return "", fmt.Errorf("API key not found for provider: %s", provider)
}
