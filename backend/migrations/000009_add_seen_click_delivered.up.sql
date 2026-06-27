-- Migration: add_seen_click_delivered
-- Version: 9
-- Description: Add seen_at, delivered_at, clicked_at tracking columns for in-app notification lifecycle

ALTER TABLE notifications ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS seen_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS clicked_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_notifications_seen_at ON notifications(seen_at);
CREATE INDEX IF NOT EXISTS idx_notifications_delivered_at ON notifications(delivered_at);
CREATE INDEX IF NOT EXISTS idx_notifications_clicked_at ON notifications(clicked_at);
