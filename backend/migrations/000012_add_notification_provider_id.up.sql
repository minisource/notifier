-- Add provider_id to notifications: explicit provider selection (no failover)
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS provider_id UUID;
CREATE INDEX IF NOT EXISTS idx_notifications_provider_id ON notifications(provider_id);
