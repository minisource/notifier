package initializer

import (
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/database"
	"gorm.io/gorm"
)

// InitDatabase initializes PostgreSQL database connection
func InitDatabase(cfg *config.Config, logger logging.Logger) *gorm.DB {
	logger.Info(logging.General, logging.Startup, "Initializing database", nil)
	db, err := database.InitDatabase(&cfg.Postgres, &cfg.Database, logger)
	if err != nil {
		logger.Fatal(logging.Postgres, logging.Startup, "Failed to initialize database", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
	}
	return db
}

// CloseDatabase closes database connection gracefully
func CloseDatabase(db *gorm.DB, logger logging.Logger) {
	database.CloseDatabase(db, logger)
}
