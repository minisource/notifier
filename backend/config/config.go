package config

import (
	"log"
	"os"
	"strconv"
	"strings"
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
	Kavenegar        KavenegarConfig
	Tracing          TracingConfig
	Digest           DigestConfig
	ProviderLogs     ProviderLogsConfig
	ProviderBalance  ProviderBalanceConfig
	DeliveryControl  DeliveryControlConfig
	TelegramGateway  TelegramGatewayConfig
}

type AuthConfig struct {
	Enabled          bool
	BaseURL          string
	ClientID         string
	ClientSecret     string
	JWTSecret        string
	Issuer           string
	Audience         string
	JWKSURL          string
	IntrospectionURL string
	ValidationMode   string
	JWKSCacheTTL     int
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
	ControlMutationRequests int // pause/resume (strict, low frequency)
	ControlReadRequests    int // status/history/held polling
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

// ProviderBalanceConfig controls the Provider Account Balance, Quota, and
// Credit Alerting feature (see design-system TODO). Safe defaults: refresh is
// periodic with jitter, stale-after marks old values without zeroing them, and
// thresholds are configurable per provider account (see balanceSettings in the
// provider config JSON). A single monetary default is NOT safe across
// providers, so thresholds live per-account; these are only process defaults.
type ProviderBalanceConfig struct {
	Enabled                bool // Master switch (PROVIDER_BALANCE_ENABLED)
	RefreshIntervalSec     int  // Default refresh interval (PROVIDER_BALANCE_REFRESH_INTERVAL_SEC)
	StaleAfterSec          int  // Mark health stale when no successful refresh in this window (PROVIDER_BALANCE_STALE_AFTER_SEC)
	RefreshTimeoutSec      int  // Per-refresh HTTP timeout (PROVIDER_BALANCE_REFRESH_TIMEOUT_SEC)
	MaxConsecutiveFailures int  // Backoff trigger after N consecutive failures (PROVIDER_BALANCE_MAX_CONSECUTIVE_FAILURES)
	RetentionDays          int  // Balance snapshot retention (PROVIDER_BALANCE_RETENTION_DAYS)
}

// DeliveryControlConfig controls the Global Outbound Delivery Pause /
// Emergency Freeze feature (see design-system TODO). The authoritative state
// is durable in the DB; the cache is only a short-TTL optimization and FAILS
// CLOSED. Release settings bound post-resume backlog so it resumes gradually.
// The remaining fields are the Critical Safety Layer: idempotency retention,
// input bounds, and stuck-worker (lease) recovery.
type DeliveryControlConfig struct {
	Enabled                   bool // Master switch (DELIVERY_CONTROL_ENABLED)
	CacheTTLSeconds           int  // How long workers cache pause state (DELIVERY_CONTROL_CACHE_TTL_SECONDS)
	ReleaseIntervalSec        int  // How often the worker releases held work (DELIVERY_CONTROL_RELEASE_INTERVAL_SECONDS)
	ReleaseBatchSize          int  // Max held deliveries released per tick (DELIVERY_CONTROL_RELEASE_BATCH_SIZE)
	MaxReasonLength           int  // Hard cap for pause/resume reason text (DELIVERY_CONTROL_MAX_REASON_LENGTH)
	MaxPauseDurationHours     int  // Reject expiresAt further out than this (DELIVERY_CONTROL_MAX_PAUSE_DURATION_HOURS, 0 = unlimited)
	IdempotencyRetentionHours int  // How long idempotency records are kept (DELIVERY_CONTROL_IDEMPOTENCY_RETENTION_HOURS)
	StuckSendingAgeSec        int  // Recover sending/processing older than this back to pending (DELIVERY_CONTROL_STUCK_SENDING_AGE_SEC)
}

// TelegramGatewayConfig controls the Telegram Gateway OTP provider adapter.
// Telegram Gateway (gatewayapi.telegram.org) is NOT the Bot API: it sends
// official verification messages (sendVerificationMessage / checkSendAbility /
// checkVerificationStatus / revokeVerificationMessage). The API token is a
// secret — it comes from the environment only, is never exposed via admin
// APIs, and is redacted from logs/traces/provider lifecycle records.
type TelegramGatewayConfig struct {
	Enabled            bool   // Master switch (TELEGRAM_GATEWAY_ENABLED)
	APIToken           string // Secret token (TELEGRAM_GATEWAY_API_TOKEN) — never logged
	BaseURL            string // API root (TELEGRAM_GATEWAY_BASE_URL), default https://gatewayapi.telegram.org
	RequestTimeoutSec  int    // Per-request overall deadline (TELEGRAM_GATEWAY_REQUEST_TIMEOUT_SECONDS)
	ConnectTimeoutSec  int    // TCP connect/TLS handshake deadline (TELEGRAM_GATEWAY_CONNECT_TIMEOUT_SECONDS)
	MaxResponseBytes   int    // Response-body read cap (TELEGRAM_GATEWAY_MAX_RESPONSE_BYTES)
	TestMode           bool   // Test mode flag (TELEGRAM_GATEWAY_TEST_MODE)
	DefaultTTL         int    // Default code TTL seconds, 30..3600 (TELEGRAM_GATEWAY_DEFAULT_TTL)
	DefaultCodeLength  int    // Default generated-code length 4..8 (TELEGRAM_GATEWAY_DEFAULT_CODE_LENGTH)
	CheckPhone         string // Optional E.164 phone for live health checks (TELEGRAM_GATEWAY_CHECK_PHONE)
}

// ProviderLogsConfig controls the durable Provider Request Lifecycle Logging
// feature (see design-system TODO: "Provider Request Lifecycle Logging").
// Defaults are safe: bodies are bounded, recipient/content is never stored in
// full, and retention is configurable per data category.
type ProviderLogsConfig struct {
	Enabled             bool // Master switch (PROVIDER_LOGS_ENABLED)
	BodyCaptureMaxBytes int  // Max captured request/response body bytes before truncation (PROVIDER_LOGS_BODY_CAPTURE_MAX_BYTES)
	BodyPreviewMaxChars int  // Max stored message preview chars (PROVIDER_LOGS_BODY_PREVIEW_MAX_CHARS)
	MetadataRetentionDays int // Full-row retention (PROVIDER_LOGS_METADATA_RETENTION_DAYS)
	BodyRetentionDays   int  // Captured body retention before purge, metadata kept (PROVIDER_LOGS_BODY_RETENTION_DAYS)
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
				AllowOrigins:     getEnv("CORS_ALLOW_ORIGINS", "http://localhost:3003,http://127.0.0.1:3003,http://localhost:3004,http://127.0.0.1:3004"),
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
				ControlMutationRequests: getEnvAsInt("RATE_LIMIT_CONTROL_MUTATION_REQUESTS", 10),
				ControlReadRequests:    getEnvAsInt("RATE_LIMIT_CONTROL_READ_REQUESTS", 120),
			},
			Auth: AuthConfig{
				Enabled:          getEnvAsBool("AUTH_ENABLED", true),
				BaseURL:          getEnv("AUTH_BASE_URL", "http://localhost:9001"),
				ClientID:         getEnv("AUTH_CLIENT_ID", ""),
				ClientSecret:     getEnv("AUTH_CLIENT_SECRET", ""),
				JWTSecret:        getEnv("AUTH_JWT_SECRET", ""),
				Issuer:           getEnv("AUTH_ISSUER", "minisource-auth"),
				Audience:         getEnv("AUTH_AUDIENCE", "minisource"),
				JWKSURL:          getEnv("AUTH_JWKS_URL", "http://localhost:9001/.well-known/jwks.json"),
				IntrospectionURL: getEnv("AUTH_INTROSPECTION_URL", "http://localhost:9001/v1/auth/introspect"),
				ValidationMode:   getEnv("AUTH_VALIDATION_MODE", "hs256"),
				JWKSCacheTTL:     getEnvAsInt("AUTH_JWKS_CACHE_TTL_SECONDS", 300),
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
			ProviderLogs: ProviderLogsConfig{
				Enabled:               getEnvAsBool("PROVIDER_LOGS_ENABLED", true),
				BodyCaptureMaxBytes:   getEnvAsInt("PROVIDER_LOGS_BODY_CAPTURE_MAX_BYTES", 8192),
				BodyPreviewMaxChars:   getEnvAsInt("PROVIDER_LOGS_BODY_PREVIEW_MAX_CHARS", 200),
				MetadataRetentionDays: getEnvAsInt("PROVIDER_LOGS_METADATA_RETENTION_DAYS", 30),
				BodyRetentionDays:     getEnvAsInt("PROVIDER_LOGS_BODY_RETENTION_DAYS", 7),
			},
			ProviderBalance: ProviderBalanceConfig{
				Enabled:                getEnvAsBool("PROVIDER_BALANCE_ENABLED", true),
				RefreshIntervalSec:     getEnvAsInt("PROVIDER_BALANCE_REFRESH_INTERVAL_SEC", 3600),
				StaleAfterSec:          getEnvAsInt("PROVIDER_BALANCE_STALE_AFTER_SEC", 21600),
				RefreshTimeoutSec:      getEnvAsInt("PROVIDER_BALANCE_REFRESH_TIMEOUT_SEC", 10),
				MaxConsecutiveFailures: getEnvAsInt("PROVIDER_BALANCE_MAX_CONSECUTIVE_FAILURES", 3),
				RetentionDays:          getEnvAsInt("PROVIDER_BALANCE_RETENTION_DAYS", 90),
			},
			DeliveryControl: DeliveryControlConfig{
				Enabled:                   getEnvAsBool("DELIVERY_CONTROL_ENABLED", true),
				CacheTTLSeconds:           getEnvAsInt("DELIVERY_CONTROL_CACHE_TTL_SECONDS", 2),
				ReleaseIntervalSec:        getEnvAsInt("DELIVERY_CONTROL_RELEASE_INTERVAL_SECONDS", 5),
				ReleaseBatchSize:          getEnvAsInt("DELIVERY_CONTROL_RELEASE_BATCH_SIZE", 50),
				MaxReasonLength:           getEnvAsInt("DELIVERY_CONTROL_MAX_REASON_LENGTH", 500),
				MaxPauseDurationHours:     getEnvAsInt("DELIVERY_CONTROL_MAX_PAUSE_DURATION_HOURS", 168),
				IdempotencyRetentionHours: getEnvAsInt("DELIVERY_CONTROL_IDEMPOTENCY_RETENTION_HOURS", 24),
				StuckSendingAgeSec:        getEnvAsInt("DELIVERY_CONTROL_STUCK_SENDING_AGE_SEC", 300),
			},
			TelegramGateway: TelegramGatewayConfig{
				Enabled:           getEnvAsBool("TELEGRAM_GATEWAY_ENABLED", false),
				APIToken:          os.Getenv("TELEGRAM_GATEWAY_API_TOKEN"),
				BaseURL:           getEnv("TELEGRAM_GATEWAY_BASE_URL", "https://gatewayapi.telegram.org"),
				RequestTimeoutSec: getEnvAsInt("TELEGRAM_GATEWAY_REQUEST_TIMEOUT_SECONDS", 15),
				ConnectTimeoutSec: getEnvAsInt("TELEGRAM_GATEWAY_CONNECT_TIMEOUT_SECONDS", 5),
				MaxResponseBytes:  getEnvAsInt("TELEGRAM_GATEWAY_MAX_RESPONSE_BYTES", 1<<20),
				TestMode:          getEnvAsBool("TELEGRAM_GATEWAY_TEST_MODE", false),
				DefaultTTL:        getEnvAsInt("TELEGRAM_GATEWAY_DEFAULT_TTL", 120),
				DefaultCodeLength: getEnvAsInt("TELEGRAM_GATEWAY_DEFAULT_CODE_LENGTH", 6),
				CheckPhone:        os.Getenv("TELEGRAM_GATEWAY_CHECK_PHONE"),
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
		if c.Auth.ValidationMode == "hs256" && c.Auth.JWTSecret == "" {
			issues = append(issues, "AUTH_JWT_SECRET is not set but AUTH_ENABLED=true with hs256 validation")
		}
		if c.Auth.ValidationMode == "jwks" && c.Auth.JWKSURL == "" {
			issues = append(issues, "AUTH_JWKS_URL is not set but AUTH_VALIDATION_MODE=jwks")
		}
		if c.Auth.ValidationMode == "introspection" && c.Auth.IntrospectionURL == "" {
			issues = append(issues, "AUTH_INTROSPECTION_URL is not set but AUTH_VALIDATION_MODE=introspection")
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

	// Delivery-control safety layer: unsafe values fail startup loudly.
	if c.DeliveryControl.Enabled {
		if c.DeliveryControl.MaxReasonLength < 0 {
			issues = append(issues, "DELIVERY_CONTROL_MAX_REASON_LENGTH must be >= 0 (0 = unlimited)")
		}
		if c.DeliveryControl.MaxPauseDurationHours < 0 {
			issues = append(issues, "DELIVERY_CONTROL_MAX_PAUSE_DURATION_HOURS must be >= 0 (0 = unlimited)")
		}
		if c.DeliveryControl.IdempotencyRetentionHours <= 0 {
			issues = append(issues, "DELIVERY_CONTROL_IDEMPOTENCY_RETENTION_HOURS must be > 0")
		}
		if c.DeliveryControl.StuckSendingAgeSec <= 0 {
			issues = append(issues, "DELIVERY_CONTROL_STUCK_SENDING_AGE_SEC must be > 0 (lease recovery)")
		}
	}
	if c.RateLimit.Enabled {
		if c.RateLimit.ControlMutationRequests <= 0 {
			issues = append(issues, "RATE_LIMIT_CONTROL_MUTATION_REQUESTS must be > 0")
		}
		if c.RateLimit.ControlReadRequests <= 0 {
			issues = append(issues, "RATE_LIMIT_CONTROL_READ_REQUESTS must be > 0")
		}
	}

	// Telegram Gateway safety checks.
	if c.TelegramGateway.Enabled {
		if c.TelegramGateway.APIToken == "" {
			issues = append(issues, "TELEGRAM_GATEWAY_API_TOKEN is not set but TELEGRAM_GATEWAY_ENABLED=true")
		}
		if c.Server.RunMode == "production" {
			base := strings.ToLower(c.TelegramGateway.BaseURL)
			if !strings.HasPrefix(base, "https://") {
				issues = append(issues, "TELEGRAM_GATEWAY_BASE_URL must be HTTPS in production")
			}
		}
		if c.TelegramGateway.DefaultTTL < 30 || c.TelegramGateway.DefaultTTL > 3600 {
			issues = append(issues, "TELEGRAM_GATEWAY_DEFAULT_TTL must be within 30..3600 seconds")
		}
		if c.TelegramGateway.DefaultCodeLength < 4 || c.TelegramGateway.DefaultCodeLength > 8 {
			issues = append(issues, "TELEGRAM_GATEWAY_DEFAULT_CODE_LENGTH must be within 4..8")
		}
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
