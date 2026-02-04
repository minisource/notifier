-- Remove tenant_id columns from notifier database tables

-- Remove indexes
DROP INDEX IF EXISTS idx_sms_templates_tenant;
DROP INDEX IF EXISTS idx_settings_key_tenant;
DROP INDEX IF EXISTS idx_settings_tenant;
DROP INDEX IF EXISTS idx_notification_logs_tenant;
DROP INDEX IF EXISTS idx_notification_preferences_user_tenant;
DROP INDEX IF EXISTS idx_notification_preferences_tenant;
DROP INDEX IF EXISTS idx_notification_templates_name_tenant;
DROP INDEX IF EXISTS idx_notification_templates_tenant;
DROP INDEX IF EXISTS idx_notifications_status_tenant;
DROP INDEX IF EXISTS idx_notifications_user_tenant;
DROP INDEX IF EXISTS idx_notifications_tenant;

-- Remove tenant_id columns
ALTER TABLE sms_templates DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE settings DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE notification_logs DROP COLUMN IF EXISTS tenant_id;
-- notification_preferences already has tenant_id in model, keep it
-- notification_templates already has tenant_id in model, keep it
-- notifications already has tenant_id in model, keep it

-- Drop tenants table
DROP TABLE IF EXISTS tenants;
