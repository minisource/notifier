-- Migration: add_idempotency_key — Rollback
DROP INDEX IF EXISTS idx_notif_idempotency_key;
ALTER TABLE notifications DROP COLUMN IF EXISTS idempotency_key;
