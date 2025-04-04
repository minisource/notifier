package config

import (
	"errors"
	"log"
	"os"

	"github.com/minisource/common_go/logging"
	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig
	// Postgres PostgresConfig
	Cors   CorsConfig
	Logger logging.LoggerConfig
	SMS    SMSConfig
}

type ServerConfig struct {
	InternalPort string
	ExternalPort string
	RunMode      string
	Name         string
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
	AllowOrigins string
}

type SMSConfig struct {
	Providers       []SMSProviderConfig
	DefualtProvider string
	NotEnabled      bool
}

type SMSProviderConfig struct {
	Provider  string
	ApiKey    string
	AccessId  string
	AccessKey string
	Sign      string
	Template  string
}

func GetConfig() *Config {
	cfgPath := getConfigPath(os.Getenv("APP_ENV"))
	v, err := LoadConfig(cfgPath, "yml")
	if err != nil {
		log.Fatalf("Error in load config %v", err)
	}

	cfg, err := ParseConfig(v)
	envPort := os.Getenv("PORT")
	if envPort != "" {
		cfg.Server.ExternalPort = envPort
		log.Printf("Set external port from environment -> %s", cfg.Server.ExternalPort)
	} else {
		cfg.Server.ExternalPort = cfg.Server.InternalPort
		log.Printf("Set external port from environment -> %s", cfg.Server.ExternalPort)
	}

	if err != nil {
		log.Fatalf("Error in parse config %v", err)
	}

	return cfg
}

func ParseConfig(v *viper.Viper) (*Config, error) {
	var cfg Config
	err := v.Unmarshal(&cfg)
	if err != nil {
		log.Printf("Unable to parse config: %v", err)
		return nil, err
	}

	return &cfg, nil
}

func LoadConfig(filename string, fileType string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType(fileType)
	v.SetConfigName(filename)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	if err != nil {
		log.Printf("Unable to read config: %v", err)
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("config file not found")
		}

		return nil, err
	}

	return v, nil
}

func getConfigPath(env string) string {
	if env == "docker" {
		return "/app/config/config-docker.yml"
	} else if env == "production" {
		return "../../config/config-production.yml"
	} else {
		return "../../config/config-development.yml"
	}
}

func (cfg *SMSConfig) GetProviderConfig(providerName string) (*SMSProviderConfig, error) {
	for _, provider := range cfg.Providers {
		if provider.Provider == providerName {
			return &provider, nil
		}
	}
	return nil, errors.New("provider not found")
}
