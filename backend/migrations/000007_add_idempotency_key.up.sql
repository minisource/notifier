-- Migration: add_idempotency_key
-- Version: 7
-- Description: Add idempotency key column to prevent duplicate notification sends

ALTER TABLE notifications ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(255) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_notif_idempotency_key 
    ON notifications(idempotency_key) WHERE idempotency_key != '';
