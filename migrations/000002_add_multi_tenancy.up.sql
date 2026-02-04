-- Add tenant_id columns to notifier database tables

-- Add tenants table (if not exists - might be shared)
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
CREATE INDEX IF NOT EXISTS idx_tenants_active ON tenants(is_active) WHERE deleted_at IS NULL;

-- notifications table already has tenant_id, ensure index
CREATE INDEX IF NOT EXISTS idx_notifications_tenant ON notifications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_tenant ON notifications(user_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status_tenant ON notifications(status, tenant_id);

-- notification_templates table already has tenant_id, ensure index
CREATE INDEX IF NOT EXISTS idx_notification_templates_tenant ON notification_templates(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notification_templates_name_tenant ON notification_templates(name, type, tenant_id);

-- notification_preferences table already has tenant_id, ensure index
CREATE INDEX IF NOT EXISTS idx_notification_preferences_tenant ON notification_preferences(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_tenant ON notification_preferences(user_id, tenant_id);

-- notification_logs table - add tenant_id if not exists
ALTER TABLE notification_logs ADD COLUMN IF NOT EXISTS tenant_id UUID;
CREATE INDEX IF NOT EXISTS idx_notification_logs_tenant ON notification_logs(tenant_id);

-- settings table - add tenant_id if not exists
ALTER TABLE settings ADD COLUMN IF NOT EXISTS tenant_id UUID;
CREATE INDEX IF NOT EXISTS idx_settings_tenant ON settings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_settings_key_tenant ON settings(key, tenant_id);

-- sms_templates table - add tenant_id if not exists
ALTER TABLE sms_templates ADD COLUMN IF NOT EXISTS tenant_id UUID;
CREATE INDEX IF NOT EXISTS idx_sms_templates_tenant ON sms_templates(tenant_id);

-- Insert default tenant if not exists
INSERT INTO tenants (id, name, slug, is_active, settings)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 
    'Default Tenant', 
    'default', 
    true, 
    '{}'
) ON CONFLICT (slug) DO NOTHING;

-- Update existing records to use default tenant
UPDATE notifications SET tenant_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11' WHERE tenant_id IS NULL;
UPDATE notification_templates SET tenant_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11' WHERE tenant_id IS NULL;
UPDATE notification_preferences SET tenant_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11' WHERE tenant_id IS NULL;
UPDATE notification_logs SET tenant_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11' WHERE tenant_id IS NULL;
UPDATE settings SET tenant_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11' WHERE tenant_id IS NULL;
UPDATE sms_templates SET tenant_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11' WHERE tenant_id IS NULL;
