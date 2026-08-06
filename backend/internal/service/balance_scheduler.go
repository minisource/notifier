package service

import (
	"context"
	"time"

	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/config"
)

// BalanceScheduler periodically refreshes provider account balance/quota data.
// It reads persisted values for dashboards; it never makes dashboard page loads
// trigger external provider calls. Refresh cadence is configurable; jitter is
// applied inside RefreshAccount.
type BalanceScheduler struct {
	cfg     *config.Config
	logger  logging.Logger
	balance *BalanceService
	stop    chan struct{}
}

// NewBalanceScheduler creates the scheduler.
func NewBalanceScheduler(cfg *config.Config, logger logging.Logger, balance *BalanceService) *BalanceScheduler {
	return &BalanceScheduler{
		cfg:     cfg,
		logger:  logger,
		balance: balance,
		stop:    make(chan struct{}),
	}
}

// Start launches the periodic refresh loop. It never blocks callers.
func (s *BalanceScheduler) Start() {
	if !s.cfg.ProviderBalance.Enabled {
		s.logger.Info(logging.General, logging.Startup, "Provider balance scheduler disabled", nil)
		return
	}
	interval := time.Duration(s.cfg.ProviderBalance.RefreshIntervalSec) * time.Second
	if interval < 60*time.Second {
		interval = 60 * time.Second
	}
	s.logger.Info(logging.General, logging.Startup, "Provider balance scheduler started", map[logging.ExtraKey]interface{}{
		logging.ExtraKey("interval"): interval.String(),
	})
	go s.loop(interval)
}

// Stop stops the scheduler.
func (s *BalanceScheduler) Stop() {
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
}

func (s *BalanceScheduler) loop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.balance.RefreshAll(context.Background())
		case <-s.stop:
			return
		}
	}
}
