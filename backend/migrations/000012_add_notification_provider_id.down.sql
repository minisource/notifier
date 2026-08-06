DROP INDEX IF EXISTS idx_notifications_provider_id;
ALTER TABLE notifications DROP COLUMN IF EXISTS provider_id;
