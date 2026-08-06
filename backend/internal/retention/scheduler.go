package retention

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-common/retention"
)

// Scheduler periodically runs enabled retention policies.
type Scheduler struct {
	policyRepo PolicyRepository
	runRepo    RunRepository
	runner     *NotifierRunner
	lock       *PGLock
	logger     logging.Logger
	stop       chan struct{}
	mu         sync.Mutex
	running    bool
}

func NewScheduler(
	policyRepo PolicyRepository,
	runRepo RunRepository,
	runner *NotifierRunner,
	lock *PGLock,
	logger logging.Logger,
) *Scheduler {
	return &Scheduler{
		policyRepo: policyRepo,
		runRepo:    runRepo,
		runner:     runner,
		lock:       lock,
		logger:     logger,
		stop:       make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()
	s.logger.Info(logging.General, logging.Startup, "Retention scheduler started", nil)
	go s.loop()
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	close(s.stop)
	s.running = false
}

func (s *Scheduler) GetRunner() *NotifierRunner { return s.runner }

func (s *Scheduler) loop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	s.tick()
	for {
		select {
		case <-ticker.C:
			s.tick()
		case <-s.stop:
			s.logger.Info(logging.General, logging.Startup, "Retention scheduler stopped", nil)
			return
		}
	}
}

func (s *Scheduler) tick() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	policies, err := s.policyRepo.ListEnabled(ctx)
	if err != nil {
		s.logger.Error(logging.General, logging.Startup, "Failed to list enabled policies", map[logging.ExtraKey]interface{}{"error": err.Error()})
		return
	}
	for _, pm := range policies {
		s.executePolicy(ctx, pm)
	}
}

func (s *Scheduler) executePolicy(ctx context.Context, pm PolicyModel) {
	policy := pm.ToDomain()
	if err := ValidatePolicy(&policy); err != nil {
		s.logger.Warn(logging.General, logging.Startup, "Skipping invalid policy", map[logging.ExtraKey]interface{}{"category": policy.Category, "error": err.Error()})
		return
	}
	lockKeyStr := LockKey(policy.Service, policy.Category)
	guard, err := s.lock.Acquire(ctx, lockKeyStr, 10*time.Minute)
	if err != nil {
		if err == retention.ErrLockHeld {
			s.logger.Debug(logging.General, logging.Startup, "Cleanup lock held", map[logging.ExtraKey]interface{}{"category": policy.Category})
			retention.CleanupLockContentionTotal.WithLabelValues(policy.Service, policy.Category).Inc()
		}
		return
	}
	defer guard.Release(ctx)

	cutoff := policy.ComputeCutoff(time.Now().UTC())
	if policy.Strategy == retention.StrategyCount || policy.Strategy == retention.StrategyHybrid {
		if countCutoff, err := s.runner.ComputeCountCutoff(ctx, policy.Category, policy.KeepLatestCount); err == nil && !countCutoff.IsZero() {
			if policy.Strategy == retention.StrategyCount || countCutoff.Before(cutoff) {
				cutoff = countCutoff
			}
		}
	}

	snapshot := retention.RunSnapshot{
		PolicyID: policy.ID, Service: policy.Service, Category: policy.Category,
		Strategy: policy.Strategy, DryRun: policy.DryRun, Cutoff: cutoff,
		KeepLatest: policy.KeepLatestCount, BatchSize: policy.EffectiveBatchSize(),
		MaxBatches: policy.EffectiveMaxBatches(), Trigger: retention.TriggerScheduled,
		RunID: uuid.New().String(), StartedAt: time.Now().UTC(),
	}

	retention.CleanupActive.WithLabelValues(policy.Service, policy.Category).Set(1)
	defer retention.CleanupActive.WithLabelValues(policy.Service, policy.Category).Set(0)

	runner, err := s.runner.NewSharedRunner(snapshot)
	if err != nil {
		s.logger.Error(logging.General, logging.Startup, "Failed to create runner", map[logging.ExtraKey]interface{}{"category": policy.Category, "error": err.Error()})
		return
	}
	result := runner.Run(ctx)

	RecordRun(ctx, s.runRepo, snapshot, result) // best-effort

	retention.CleanupRunsTotal.WithLabelValues(policy.Service, policy.Category, string(result.Result), string(snapshot.Trigger)).Inc()
	retention.CleanupDeletedRecordsTotal.WithLabelValues(policy.Service, policy.Category).Add(float64(result.DeletedCount))
	dur := result.EndedAt.Sub(result.StartedAt).Seconds()
	retention.CleanupDurationSeconds.WithLabelValues(policy.Service, policy.Category, string(result.Result)).Observe(dur)
	if result.Result == retention.ResultSuccess || result.Result == retention.ResultPartial {
		retention.CleanupLastSuccessTimestamp.WithLabelValues(policy.Service, policy.Category).Set(float64(result.EndedAt.Unix()))
	}

	now := time.Now().UTC()
	pm.LastRunAt = &now
	_ = s.policyRepo.Update(ctx, &pm)
}

// ExecuteManual runs a cleanup policy immediately (manual trigger).
func (s *Scheduler) ExecuteManual(ctx context.Context, pm PolicyModel) (*RunRecord, error) {
	policy := pm.ToDomain()
	if err := ValidatePolicy(&policy); err != nil {
		return nil, err
	}
	lockKeyStr := LockKey(policy.Service, policy.Category)
	guard, err := s.lock.Acquire(ctx, lockKeyStr, 10*time.Minute)
	if err != nil {
		return nil, err
	}
	defer guard.Release(ctx)

	cutoff := policy.ComputeCutoff(time.Now().UTC())
	if policy.Strategy == retention.StrategyCount || policy.Strategy == retention.StrategyHybrid {
		if countCutoff, err := s.runner.ComputeCountCutoff(ctx, policy.Category, policy.KeepLatestCount); err == nil && !countCutoff.IsZero() {
			if policy.Strategy == retention.StrategyCount || countCutoff.Before(cutoff) {
				cutoff = countCutoff
			}
		}
	}

	snapshot := retention.RunSnapshot{
		PolicyID: policy.ID, Service: policy.Service, Category: policy.Category,
		Strategy: policy.Strategy, DryRun: policy.DryRun, Cutoff: cutoff,
		KeepLatest: policy.KeepLatestCount, BatchSize: policy.EffectiveBatchSize(),
		MaxBatches: policy.EffectiveMaxBatches(), Trigger: retention.TriggerManual,
		RunID: uuid.New().String(), StartedAt: time.Now().UTC(),
	}

	retention.CleanupActive.WithLabelValues(policy.Service, policy.Category).Set(1)
	defer retention.CleanupActive.WithLabelValues(policy.Service, policy.Category).Set(0)

	runner, err := s.runner.NewSharedRunner(snapshot)
	if err != nil {
		return nil, err
	}
	result := runner.Run(ctx)
	record, err := RecordRun(ctx, s.runRepo, snapshot, result)

	retention.CleanupRunsTotal.WithLabelValues(policy.Service, policy.Category, string(result.Result), string(snapshot.Trigger)).Inc()
	retention.CleanupDeletedRecordsTotal.WithLabelValues(policy.Service, policy.Category).Add(float64(result.DeletedCount))
	dur := result.EndedAt.Sub(result.StartedAt).Seconds()
	retention.CleanupDurationSeconds.WithLabelValues(policy.Service, policy.Category, string(result.Result)).Observe(dur)

	now := time.Now().UTC()
	pm.LastRunAt = &now
	_ = s.policyRepo.Update(ctx, &pm)
	return record, err
}
