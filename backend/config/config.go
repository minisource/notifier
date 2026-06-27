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
	RateLimit RateLimitConfig
	Logger    logging.LoggerConfig
	Worker    WorkerConfig
	GRPC      GRPCConfig
	Auth      AuthConfig
	Database  DatabaseConfig
	Kavenegar KavenegarConfig
	Tracing   TracingConfig
	Digest    DigestConfig
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
	AutoMigrate   bool // GORM AutoMigrate (separate from SQL migrations)
}

type WorkerConfig struct {
	NumWorkers     int
	QueueSize      int
	RetryMaxDelay  int
	RetryBaseDelay int
	PollEnabled    bool // Enable periodic DB polling for pending notifications
	PollInterval   int  // Polling interval in seconds
}

type CorsConfig struct {
	AllowOrigins string
	AllowMethods string
	AllowHeaders string
	AllowCredentials bool
}

type RateLimitConfig struct {
	Enabled                bool
	Requests               int
	WindowSeconds          int
	ProviderTestRequests   int
	NotificationCreateRequests int
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

type DigestConfig struct {
	Enabled     bool
	Interval    int // How often to process digests (seconds)
	BatchSize   int // Max notifications per digest
	MaxBodyLen  int // Max length of each item body in digest
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
				PollEnabled:    getEnvAsBool("WORKER_POLL_ENABLED", true),
				PollInterval:   getEnvAsInt("WORKER_POLL_INTERVAL", 15),
			},
			Cors: CorsConfig{
				AllowOrigins:     getEnv("CORS_ALLOW_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000,http://localhost:3001,http://127.0.0.1:3001"),
				AllowMethods:     getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"),
				AllowHeaders:     getEnv("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization,X-Request-Id,X-Tenant-Id"),
				AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", false),
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
			RateLimit: RateLimitConfig{
				Enabled:                getEnvAsBool("RATE_LIMIT_ENABLED", true),
				Requests:               getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
				WindowSeconds:          getEnvAsInt("RATE_LIMIT_WINDOW_SECONDS", 60),
				ProviderTestRequests:   getEnvAsInt("RATE_LIMIT_PROVIDER_TEST_REQUESTS", 10),
				NotificationCreateRequests: getEnvAsInt("RATE_LIMIT_NOTIFICATION_CREATE_REQUESTS", 30),
			},
			Auth: AuthConfig{
				Enabled:      getEnvAsBool("AUTH_ENABLED", true),
				BaseURL:      getEnv("AUTH_BASE_URL", "http://localhost:9001"),
				ClientID:     getEnv("AUTH_CLIENT_ID", ""),
				ClientSecret: getEnv("AUTH_CLIENT_SECRET", ""),
				JWTSecret:    getEnv("AUTH_JWT_SECRET", ""),
			},
			Database: DatabaseConfig{
				RunMigrations: getEnvAsBool("DB_RUN_MIGRATIONS", false),
				RunSeedData:   getEnvAsBool("DB_RUN_SEED_DATA", false),
				AutoMigrate:   getEnvAsBool("DB_AUTO_MIGRATE", false),
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
			Digest: DigestConfig{
				Enabled:    getEnvAsBool("DIGEST_ENABLED", true),
				Interval:   getEnvAsInt("DIGEST_INTERVAL", 60),
				BatchSize:  getEnvAsInt("DIGEST_BATCH_SIZE", 50),
				MaxBodyLen: getEnvAsInt("DIGEST_MAX_BODY_LEN", 200),
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

// Validate checks critical configuration values and returns warnings/errors.
func (c *Config) Validate() []string {
	var issues []string

	if c.Server.InternalPort == "" {
		issues = append(issues, "SERVER_INTERNAL_PORT is not set")
	}
	if c.Server.ExternalPort == "" {
		issues = append(issues, "SERVER_EXTERNAL_PORT is not set")
	}
	if c.Postgres.Host == "" {
		issues = append(issues, "POSTGRES_HOST is not set")
	}
	if c.Postgres.Port == "" {
		issues = append(issues, "POSTGRES_PORT is not set")
	}
	if c.Postgres.User == "" {
		issues = append(issues, "POSTGRES_USER is not set")
	}
	if c.Postgres.DbName == "" {
		issues = append(issues, "POSTGRES_DBNAME is not set")
	}

	if c.Auth.Enabled {
		if c.Auth.JWTSecret == "" {
			issues = append(issues, "AUTH_JWT_SECRET is not set but AUTH_ENABLED=true")
		}
	}

	if c.Cors.AllowOrigins == "" {
		issues = append(issues, "CORS_ALLOWED_ORIGINS is not set — API will be inaccessible from browsers")
	}

	if c.Worker.NumWorkers <= 0 {
		issues = append(issues, "WORKER_NUM_WORKERS must be > 0")
	}
	if c.Worker.QueueSize <= 0 {
		issues = append(issues, "WORKER_QUEUE_SIZE must be > 0")
	}

	return issues
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
