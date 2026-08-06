package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/notifier/config"
	"github.com/minisource/notifier/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBalanceRepo is an in-memory ProviderBalanceRepository used to verify the
// service's alert dedup/recovery/state logic without a database.
type fakeBalanceRepo struct {
	snapshots []*models.ProviderBalanceSnapshot
	health    map[uuid.UUID]*models.ProviderAccountHealth
	alerts    []*models.ProviderCreditAlert
	createCalls int
}

func newFakeBalanceRepo() *fakeBalanceRepo {
	return &fakeBalanceRepo{health: map[uuid.UUID]*models.ProviderAccountHealth{}}
}

func (f *fakeBalanceRepo) CreateSnapshot(_ context.Context, s *models.ProviderBalanceSnapshot) error {
	f.snapshots = append(f.snapshots, s)
	return nil
}
func (f *fakeBalanceRepo) ListSnapshots(_ context.Context, _ uuid.UUID, _, _ *time.Time, _ int) ([]*models.ProviderBalanceSnapshot, error) {
	return f.snapshots, nil
}
func (f *fakeBalanceRepo) GetLatestSnapshot(_ context.Context, _ uuid.UUID) (*models.ProviderBalanceSnapshot, error) {
	if len(f.snapshots) == 0 {
		return nil, nil
	}
	return f.snapshots[len(f.snapshots)-1], nil
}
func (f *fakeBalanceRepo) UpsertHealth(_ context.Context, h *models.ProviderAccountHealth) error {
	f.health[h.ProviderID] = h
	return nil
}
func (f *fakeBalanceRepo) GetHealth(_ context.Context, pid uuid.UUID) (*models.ProviderAccountHealth, error) {
	return f.health[pid], nil
}
func (f *fakeBalanceRepo) ListHealth(_ context.Context, _ *uuid.UUID) ([]*models.ProviderAccountHealth, error) {
	out := make([]*models.ProviderAccountHealth, 0, len(f.health))
	for _, h := range f.health {
		out = append(out, h)
	}
	return out, nil
}
func (f *fakeBalanceRepo) CreateAlert(_ context.Context, a *models.ProviderCreditAlert) error {
	f.alerts = append(f.alerts, a)
	f.createCalls++
	return nil
}
func (f *fakeBalanceRepo) UpdateAlert(_ context.Context, a *models.ProviderCreditAlert) error {
	for i, x := range f.alerts {
		if x.ID == a.ID {
			f.alerts[i] = a
			break
		}
	}
	return nil
}
func (f *fakeBalanceRepo) GetActiveAlert(_ context.Context, pid uuid.UUID, alertType string) (*models.ProviderCreditAlert, error) {
	for _, a := range f.alerts {
		if a.ProviderID == pid && a.AlertType == alertType && a.Status == models.AlertStatusActive {
			return a, nil
		}
	}
	return nil, nil
}
func (f *fakeBalanceRepo) ListAlerts(_ context.Context, _ uuid.UUID, _ string, _ int) ([]*models.ProviderCreditAlert, error) {
	return f.alerts, nil
}
func (f *fakeBalanceRepo) ListAllAlerts(_ context.Context, _ string, _ *uuid.UUID, _ int) ([]*models.ProviderCreditAlert, error) {
	return f.alerts, nil
}
func (f *fakeBalanceRepo) ResolveAlertsForType(_ context.Context, pid uuid.UUID, alertType, reason string) (int64, error) {
	var n int64
	for _, a := range f.alerts {
		if a.ProviderID == pid && a.AlertType == alertType && a.Status == models.AlertStatusActive {
			a.Status = models.AlertStatusResolved
			a.ResolvedReason = reason
			now := time.Now()
			a.ResolvedAt = &now
			n++
		}
	}
	return n, nil
}
func (f *fakeBalanceRepo) DeleteExpiredSnapshots(_ context.Context, _ time.Time) (int64, error) { return 0, nil }

func testBalanceService(fake *fakeBalanceRepo) *BalanceService {
	cfg := &config.Config{}
	cfg.ProviderBalance.RefreshTimeoutSec = 5
	cfg.ProviderBalance.StaleAfterSec = 3600
	cfg.ProviderBalance.RefreshIntervalSec = 3600
	svc := &BalanceService{cfg: cfg, balanceRepo: fake}
	return svc
}

func TestEvaluateHealthAndAlerts_ZeroIsExhaustedNotUnknown(t *testing.T) {
	fake := newFakeBalanceRepo()
	svc := testBalanceService(fake)
	pid := uuid.New()
	health := &models.ProviderAccountHealth{ProviderID: pid, Provider: "kavenegar"}
	bv := 0.0
	health.BalanceValue = &bv

	svc.evaluateHealthAndAlerts(context.Background(), health,
		&models.ProviderBalanceSnapshot{ProviderID: pid, CapabilityMode: models.BalanceCapabilityAutomatic, RefreshStatus: "success"},
		DefaultBalanceSettings())

	assert.Equal(t, models.HealthLevelExhausted, health.HealthLevel)
	assert.Equal(t, models.AlertTypeExhausted, health.LatestAlertLevel)
	active, _ := fake.GetActiveAlert(context.Background(), pid, models.AlertTypeExhausted)
	require.NotNil(t, active)
}

func TestEvaluateHealthAndAlerts_WarningThreshold(t *testing.T) {
	fake := newFakeBalanceRepo()
	svc := testBalanceService(fake)
	pid := uuid.New()
	health := &models.ProviderAccountHealth{ProviderID: pid, Provider: "kavenegar"}
	warn := 100.0
	crit := 20.0
	bv := 50.0
	health.BalanceValue = &bv
	settings := DefaultBalanceSettings()
	settings.WarningThreshold = &warn
	settings.CriticalThreshold = &crit

	svc.evaluateHealthAndAlerts(context.Background(), health,
		&models.ProviderBalanceSnapshot{ProviderID: pid, CapabilityMode: models.BalanceCapabilityAutomatic, RefreshStatus: "success"},
		settings)

	assert.Equal(t, models.HealthLevelWarning, health.HealthLevel)
	assert.Equal(t, models.AlertTypeWarning, health.LatestAlertLevel)
}

func TestEvaluateHealthAndAlerts_CriticalSupersedesWarning(t *testing.T) {
	fake := newFakeBalanceRepo()
	svc := testBalanceService(fake)
	pid := uuid.New()
	health := &models.ProviderAccountHealth{ProviderID: pid, Provider: "kavenegar"}
	warn := 100.0
	crit := 20.0
	bv := 5.0
	health.BalanceValue = &bv
	settings := DefaultBalanceSettings()
	settings.WarningThreshold = &warn
	settings.CriticalThreshold = &crit

	svc.evaluateHealthAndAlerts(context.Background(), health,
		&models.ProviderBalanceSnapshot{ProviderID: pid, CapabilityMode: models.BalanceCapabilityAutomatic, RefreshStatus: "success"},
		settings)

	assert.Equal(t, models.HealthLevelCritical, health.HealthLevel)
	assert.Equal(t, models.AlertTypeCritical, health.LatestAlertLevel)
}

func TestEvaluateHealthAndAlerts_DedupNoSpam(t *testing.T) {
	fake := newFakeBalanceRepo()
	svc := testBalanceService(fake)
	pid := uuid.New()
	warn := 100.0
	bv := 50.0
	settings := DefaultBalanceSettings()
	settings.WarningThreshold = &warn

	for i := 0; i < 5; i++ {
		health := &models.ProviderAccountHealth{ProviderID: pid, Provider: "kavenegar"}
		health.BalanceValue = &bv
		svc.evaluateHealthAndAlerts(context.Background(), health,
			&models.ProviderBalanceSnapshot{ProviderID: pid, CapabilityMode: models.BalanceCapabilityAutomatic, RefreshStatus: "success"},
			settings)
	}

	// Only one warning alert should exist after 5 identical evaluations.
	active, _ := fake.GetActiveAlert(context.Background(), pid, models.AlertTypeWarning)
	require.NotNil(t, active)
	count := 0
	for _, a := range fake.alerts {
		if a.AlertType == models.AlertTypeWarning {
			count++
		}
	}
	assert.Equal(t, 1, count)
	assert.GreaterOrEqual(t, active.RepeatCount, 4, "repeat bumps, no new alerts")
}

func TestEvaluateHealthAndAlerts_RecoveryResolvesSource(t *testing.T) {
	fake := newFakeBalanceRepo()
	svc := testBalanceService(fake)
	pid := uuid.New()
	warn := 100.0
	crit := 20.0
	low := 50.0
	high := 500.0
	settings := DefaultBalanceSettings()
	settings.WarningThreshold = &warn
	settings.CriticalThreshold = &crit

	// Trigger warning.
	h1 := &models.ProviderAccountHealth{ProviderID: pid, Provider: "kavenegar"}
	h1.BalanceValue = &low
	svc.evaluateHealthAndAlerts(context.Background(), h1,
		&models.ProviderBalanceSnapshot{ProviderID: pid, CapabilityMode: models.BalanceCapabilityAutomatic, RefreshStatus: "success"}, settings)
	require.NoError(t, fake.UpsertHealth(context.Background(), h1))
	require.NotNil(t, fake.health[pid])

	// Recover — warning must be resolved and health back to healthy.
	h2 := &models.ProviderAccountHealth{ProviderID: pid, Provider: "kavenegar"}
	h2.BalanceValue = &high
	svc.evaluateHealthAndAlerts(context.Background(), h2,
		&models.ProviderBalanceSnapshot{ProviderID: pid, CapabilityMode: models.BalanceCapabilityAutomatic, RefreshStatus: "success"}, settings)

	assert.Equal(t, models.HealthLevelHealthy, h2.HealthLevel)
	assert.Equal(t, "", h2.LatestAlertLevel)
	active, _ := fake.GetActiveAlert(context.Background(), pid, models.AlertTypeWarning)
	assert.Nil(t, active, "warning must be resolved after recovery")
}

func TestEvaluateHealthAndAlerts_FailurePreservesLastValue(t *testing.T) {
	fake := newFakeBalanceRepo()
	svc := testBalanceService(fake)
	pid := uuid.New()
	bv := 500.0
	lastSuccess := time.Now().Add(-time.Minute)
	health := &models.ProviderAccountHealth{
		ProviderID:              pid,
		Provider:                "kavenegar",
		BalanceValue:            &bv,
		LastSuccessfulRefreshAt: &lastSuccess,
	}

	// A refresh failure must NOT zero the balance.
	svc.evaluateHealthAndAlerts(context.Background(), health,
		&models.ProviderBalanceSnapshot{ProviderID: pid, CapabilityMode: models.BalanceCapabilityAutomatic, RefreshStatus: "failed", ErrorKind: "network"},
		DefaultBalanceSettings())

	require.NotNil(t, health.BalanceValue)
	assert.Equal(t, 500.0, *health.BalanceValue, "last known balance preserved on failure")
	assert.Equal(t, models.HealthLevelUnavailable, health.HealthLevel)
	active, _ := fake.GetActiveAlert(context.Background(), pid, models.AlertTypeRefreshFailed)
	require.NotNil(t, active, "refresh failure alert created")
}

// TestEvaluateHealthAndAlerts_RefreshFailedResolvedOnRecovery guards the fix
// where a refresh_failed alert stayed active forever after a successful
// refresh: recovery must also resolve refresh_failed alerts.
func TestEvaluateHealthAndAlerts_RefreshFailedResolvedOnRecovery(t *testing.T) {
	fake := newFakeBalanceRepo()
	svc := testBalanceService(fake)
	pid := uuid.New()
	bv := 500.0
	lastSuccess := time.Now().Add(-time.Minute)

	// 1) A failure creates the refresh_failed alert.
	h1 := &models.ProviderAccountHealth{ProviderID: pid, Provider: "kavenegar", BalanceValue: &bv, LastSuccessfulRefreshAt: &lastSuccess}
	svc.evaluateHealthAndAlerts(context.Background(), h1,
		&models.ProviderBalanceSnapshot{ProviderID: pid, CapabilityMode: models.BalanceCapabilityAutomatic, RefreshStatus: "failed", ErrorKind: "network"},
		DefaultBalanceSettings())
	active, _ := fake.GetActiveAlert(context.Background(), pid, models.AlertTypeRefreshFailed)
	require.NotNil(t, active, "refresh_failed alert must be created on failure")

	// 2) A subsequent success must resolve it.
	h2 := &models.ProviderAccountHealth{ProviderID: pid, Provider: "kavenegar"}
	h2.BalanceValue = &bv
	svc.evaluateHealthAndAlerts(context.Background(), h2,
		&models.ProviderBalanceSnapshot{ProviderID: pid, CapabilityMode: models.BalanceCapabilityAutomatic, RefreshStatus: "success"},
		DefaultBalanceSettings())

	active2, _ := fake.GetActiveAlert(context.Background(), pid, models.AlertTypeRefreshFailed)
	assert.Nil(t, active2, "refresh_failed alert must be resolved after a successful refresh")
}

func TestParseBalanceSettings(t *testing.T) {
	cfg := `{"provider":"kavenegar","apiKey":"x","balanceSettings":{"enabled":true,"warningThreshold":100,"criticalThreshold":20}}`
	s := ParseBalanceSettings(cfg)
	assert.True(t, s.Enabled)
	require.NotNil(t, s.WarningThreshold)
	require.NotNil(t, s.CriticalThreshold)
	assert.Equal(t, 100.0, *s.WarningThreshold)
	assert.Equal(t, 20.0, *s.CriticalThreshold)
}

func TestParseBalanceSettings_Defaults(t *testing.T) {
	s := ParseBalanceSettings(`{"provider":"kavenegar"}`)
	assert.True(t, s.Enabled)
	assert.Nil(t, s.WarningThreshold, "no invented monetary default")
	assert.Nil(t, s.CriticalThreshold)
}
