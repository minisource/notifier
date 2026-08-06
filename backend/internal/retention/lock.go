package retention

import (
	"context"
	"fmt"
	"time"

	"github.com/minisource/go-common/retention"
	"gorm.io/gorm"
)

// PGLock implements retention.DistributedLock using PostgreSQL advisory locks.
type PGLock struct {
	db *gorm.DB
}

func NewPGLock(db *gorm.DB) *PGLock {
	return &PGLock{db: db}
}

func lockKey(key string) int64 {
	var h int64
	for _, c := range key {
		h = h*31 + int64(c)
	}
	return h
}

func (l *PGLock) Acquire(ctx context.Context, key string, ttl time.Duration) (retention.LockGuard, error) {
	k := lockKey(key)
	var acquired bool
	err := l.db.WithContext(ctx).Raw("SELECT pg_try_advisory_lock(?)", k).Scan(&acquired).Error
	if err != nil {
		return nil, fmt.Errorf("pg advisory lock acquire: %w", err)
	}
	if !acquired {
		return nil, retention.ErrLockHeld
	}
	return &pgLockGuard{db: l.db, key: k, lockKey: key}, nil
}

func (l *PGLock) IsHeld(ctx context.Context, key string) (bool, error) {
	k := lockKey(key)
	var acquired bool
	err := l.db.WithContext(ctx).Raw("SELECT pg_try_advisory_lock(?)", k).Scan(&acquired).Error
	if err != nil {
		return false, err
	}
	if acquired {
		_ = l.db.WithContext(ctx).Exec("SELECT pg_advisory_unlock(?)", k).Error
		return false, nil
	}
	return true, nil
}

type pgLockGuard struct {
	db      *gorm.DB
	key     int64
	lockKey string
}

func (g *pgLockGuard) Release(ctx context.Context) error {
	return g.db.WithContext(ctx).Exec("SELECT pg_advisory_unlock(?)", g.key).Error
}

func (g *pgLockGuard) Key() string { return g.lockKey }

func LockKey(service, category string) string {
	return fmt.Sprintf("retention:%s:%s", service, category)
}

var _ retention.DistributedLock = (*PGLock)(nil)
