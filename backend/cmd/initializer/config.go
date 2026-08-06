package initializer

import (
	"context"

	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-common/metrics"
	commonTracing "github.com/minisource/go-common/tracing"
	"github.com/minisource/notifier/config"
)

// InitConfig loads configuration from environment
func InitConfig() *config.Config {
	return config.GetConfig()
}

// InitLogger creates and configures the logger
func InitLogger(cfg *config.Config) logging.Logger {
	logger := logging.NewLogger(&cfg.Logger)
	logger.Info(logging.General, logging.Startup, "Starting Notifier Service", map[logging.ExtraKey]interface{}{
		"version": "2.0.0",
	})
	return logger
}

// InitMetrics initializes Prometheus metrics
func InitMetrics() {
	metrics.InitMetrics()
}

// InitTracing initializes OpenTelemetry tracing with OTLP (go-common)
func InitTracing(cfg *config.Config, logger logging.Logger) *commonTracing.Tracer {
	tracingCfg := commonTracing.LoadConfigFromEnv()

	// Map old config if env vars not present
	if !tracingCfg.Enabled && cfg.Tracing.Enabled {
		tracingCfg.Enabled = cfg.Tracing.Enabled
		tracingCfg.ServiceName = cfg.Tracing.ServiceName
	}

	if !tracingCfg.Enabled {
		logger.Info(logging.General, logging.Startup, "Tracing disabled or not configured", nil)
		return nil
	}

	tp, err := commonTracing.InitTracer(context.Background(), tracingCfg)
	if err != nil {
		logger.Warn(logging.General, logging.Startup, "Failed to initialize tracing, continuing without it", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return nil
	}

	logger.Info(logging.General, logging.Startup, "Tracing initialized with OTLP (Tempo)", map[logging.ExtraKey]interface{}{
		"collectorURL": tracingCfg.CollectorURL,
	})

	return tp
}

// ShutdownTracing gracefully shuts down the tracer provider
func ShutdownTracing(tp *commonTracing.Tracer, logger logging.Logger) {
	if tp == nil {
		return
	}

	ctx := context.Background()
	if err := tp.Shutdown(ctx); err != nil {
		logger.Error(logging.General, logging.Startup, "Error shutting down tracer", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
	}
}

// InitTranslator initializes i18n translator
func InitTranslator(logger logging.Logger) {
	translator := i18n.GetTranslator()
	if err := translator.LoadTranslations(); err != nil {
		logger.Error(logging.General, logging.Startup, "Failed to load translations", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
	} else {
		logger.Info(logging.General, logging.Startup, "Translations loaded successfully", nil)
	}
}