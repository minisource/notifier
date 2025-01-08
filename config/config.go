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
	Password PasswordConfig
	Cors     CorsConfig
	Logger   logging.LoggerConfig
	SMS      SMSConfig
}

type SMSConfig struct {
	Provider string
	ApiKey   string
	NotSendSms  bool
}

// type PostgresConfig struct {
// 	Host            string
// 	Port            string
// 	User            string
// 	Password        string
// 	DbName          string
// 	SSLMode         string
// 	MaxIdleConns    int
// 	MaxOpenConns    int
// 	ConnMaxLifetime time.Duration
// }

type ServerConfig struct {
	InternalPort string
	ExternalPort string
	RunMode      string
}

type CorsConfig struct {
	AllowOrigins string
}

type PasswordConfig struct {
	IncludeChars     bool
	IncludeDigits    bool
	MinLength        int
	MaxLength        int
	IncludeUppercase bool
	IncludeLowercase bool
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
		return "/config/config-docker"
	} else if env == "production" {
		return "/config/config-production"
	} else {
		return "/config/config-development"
	}
}
