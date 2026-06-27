-- Rollback audit and security features
-- Migration: 000003_add_audit_security

-- Drop triggers
DROP TRIGGER IF EXISTS update_templates_updated_at ON notification_templates;
DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_templates_tenant_type;
DROP INDEX IF EXISTS idx_notifications_tenant_status;
DROP INDEX IF EXISTS idx_notifications_tenant_user;

-- Drop policies
DROP POLICY IF EXISTS tenant_isolation_audit ON audit_logs;
DROP POLICY IF EXISTS tenant_isolation_preferences ON notification_preferences;
DROP POLICY IF EXISTS tenant_isolation_templates ON notification_templates;
DROP POLICY IF EXISTS tenant_isolation_notifications ON notifications;

-- Disable RLS
ALTER TABLE audit_logs DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_preferences DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_templates DISABLE ROW LEVEL SECURITY;
ALTER TABLE notifications DISABLE ROW LEVEL SECURITY;

-- Drop table
DROP TABLE IF EXISTS audit_logs;

-- Remove audit columns
ALTER TABLE notification_templates DROP COLUMN IF EXISTS updated_by;
ALTER TABLE notification_templates DROP COLUMN IF EXISTS created_by;

ALTER TABLE notifications DROP COLUMN IF EXISTS updated_by;
ALTER TABLE notifications DROP COLUMN IF EXISTS created_by;
