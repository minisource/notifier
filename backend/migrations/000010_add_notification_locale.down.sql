-- Migration: add_notification_locale — Rollback

DROP INDEX IF EXISTS idx_notifications_locale;
ALTER TABLE notifications DROP COLUMN IF EXISTS locale;
