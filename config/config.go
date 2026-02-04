package config

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/minisource/go-common/logging"
)

var (
	cfg  *Config
	once sync.Once
)

type Config struct {
	Server    ServerConfig
	Postgres  PostgresConfig
	Cors      CorsConfig
	Logger    logging.LoggerConfig
	Worker    WorkerConfig
	GRPC      GRPCConfig
	Auth      AuthConfig
	Database  DatabaseConfig
	Kavenegar KavenegarConfig
	Tracing   TracingConfig
}

type AuthConfig struct {
	Enabled      bool
	BaseURL      string
	ClientID     string
	ClientSecret string
	JWTSecret    string
}

type ServerConfig struct {
	InternalPort string
	ExternalPort string
	RunMode      string
	Name         string
}

type PostgresConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DbName          string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
}

type DatabaseConfig struct {
	RunMigrations bool
	RunSeedData   bool
}

type WorkerConfig struct {
	NumWorkers     int
	QueueSize      int
	RetryMaxDelay  int
	RetryBaseDelay int
}

type CorsConfig struct {
	AllowOrigins string
}

type GRPCConfig struct {
	Port    string
	Enabled bool
}

// KavenegarConfig holds Kavenegar SMS provider configuration
type KavenegarConfig struct {
	Enabled  bool
	APIKey   string
	Template string // Template name for lookup API (e.g., "verify")
}

type TracingConfig struct {
	Enabled     bool
	JaegerURL   string
	ServiceName string
}

func GetConfig() *Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file not found, using environment variables")
		}

		cfg = &Config{
			Server: ServerConfig{
				InternalPort: getEnv("SERVER_INTERNAL_PORT", "9002"),
				ExternalPort: getEnv("SERVER_EXTERNAL_PORT", "9002"),
				RunMode:      getEnv("SERVER_RUN_MODE", "development"),
				Name:         getEnv("SERVER_NAME", "Notifier"),
			},
			Postgres: PostgresConfig{
				Host:            getEnv("POSTGRES_HOST", "localhost"),
				Port:            getEnv("POSTGRES_PORT", "5432"),
				User:            getEnv("POSTGRES_USER", "postgres"),
				Password:        getEnv("POSTGRES_PASSWORD", "postgres"),
				DbName:          getEnv("POSTGRES_DBNAME", "notifier_db"),
				SSLMode:         getEnv("POSTGRES_SSLMODE", "disable"),
				MaxIdleConns:    getEnvAsInt("POSTGRES_MAX_IDLE_CONNS", 10),
				MaxOpenConns:    getEnvAsInt("POSTGRES_MAX_OPEN_CONNS", 100),
				ConnMaxLifetime: getEnvAsInt("POSTGRES_CONN_MAX_LIFETIME", 60),
			},
			Worker: WorkerConfig{
				NumWorkers:     getEnvAsInt("WORKER_NUM_WORKERS", 10),
				QueueSize:      getEnvAsInt("WORKER_QUEUE_SIZE", 1000),
				RetryMaxDelay:  getEnvAsInt("WORKER_RETRY_MAX_DELAY", 300),
				RetryBaseDelay: getEnvAsInt("WORKER_RETRY_BASE_DELAY", 5),
			},
			Cors: CorsConfig{
				AllowOrigins: getEnv("CORS_ALLOW_ORIGINS", "*"),
			},
			Logger: logging.LoggerConfig{
				FilePath:    getEnv("LOGGER_FILE_PATH", "logs/notifier.log"),
				Encoding:    getEnv("LOGGER_ENCODING", "json"),
				Level:       getEnv("LOGGER_LEVEL", "debug"),
				Logger:      getEnv("LOGGER_TYPE", "zap"),
				ConsoleOnly: getEnvAsBool("LOGGER_CONSOLE_ONLY", false),
			},
			GRPC: GRPCConfig{
				Port:    getEnv("GRPC_PORT", "9003"),
				Enabled: getEnvAsBool("GRPC_ENABLED", true),
			},
			Auth: AuthConfig{
				Enabled:      getEnvAsBool("AUTH_ENABLED", true),
				BaseURL:      getEnv("AUTH_BASE_URL", "http://localhost:9001"),
				ClientID:     getEnv("AUTH_CLIENT_ID", ""),
				ClientSecret: getEnv("AUTH_CLIENT_SECRET", ""),
				JWTSecret:    getEnv("AUTH_JWT_SECRET", ""),
			},
			Database: DatabaseConfig{
				RunMigrations: getEnvAsBool("DB_RUN_MIGRATIONS", true),
				RunSeedData:   getEnvAsBool("DB_RUN_SEED_DATA", true),
			},
			Kavenegar: KavenegarConfig{
				Enabled:  getEnvAsBool("KAVENEGAR_ENABLED", false),
				APIKey:   getEnv("KAVENEGAR_API_KEY", ""),
				Template: getEnv("KAVENEGAR_TEMPLATE", "verify"),
			},
			Tracing: TracingConfig{
				Enabled:     getEnvAsBool("TRACING_ENABLED", false),
				JaegerURL:   getEnv("JAEGER_URL", "http://localhost:14268/api/traces"),
				ServiceName: getEnv("TRACING_SERVICE_NAME", "notifier-service"),
			},
		}

		if envPort := os.Getenv("PORT"); envPort != "" {
			cfg.Server.ExternalPort = envPort
			log.Printf("Set external port from PORT environment variable: %s", cfg.Server.ExternalPort)
		}

		log.Printf("Configuration loaded successfully")
		log.Printf("Server will run on port: %s", cfg.Server.InternalPort)
		if cfg.GRPC.Enabled {
			log.Printf("gRPC server will run on port: %s", cfg.GRPC.Port)
		}
	})

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: invalid integer value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("Warning: invalid boolean value for %s, using default: %t", key, defaultValue)
		return defaultValue
	}
	return value
}
