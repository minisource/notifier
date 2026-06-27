-- Migration: add_seen_click_delivered — Rollback

DROP INDEX IF EXISTS idx_notifications_seen_at;
DROP INDEX IF EXISTS idx_notifications_delivered_at;
DROP INDEX IF EXISTS idx_notifications_clicked_at;

ALTER TABLE notifications DROP COLUMN IF EXISTS clicked_at;
ALTER TABLE notifications DROP COLUMN IF EXISTS seen_at;
ALTER TABLE notifications DROP COLUMN IF EXISTS delivered_at;
