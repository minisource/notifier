-- Migration: add_template_locale_key
-- Version: 6
-- Description: Add key and locale columns to notification_templates for programmatic lookup and i18n support

-- Add key column for programmatic template lookup (e.g., "auth.otp.sms", "divipay.debt_reminder")
ALTER TABLE notification_templates ADD COLUMN IF NOT EXISTS key VARCHAR(255);

-- Create unique index for key lookup (null-safe: only applies when key IS NOT NULL)
CREATE UNIQUE INDEX IF NOT EXISTS idx_template_key_tenant 
    ON notification_templates(tenant_id, key) WHERE deleted_at IS NULL AND key IS NOT NULL;

-- Add locale column for i18n support (default: 'en')
ALTER TABLE notification_templates ADD COLUMN IF NOT EXISTS locale VARCHAR(10) NOT NULL DEFAULT 'en';

-- Migrate existing templates: set key = name for rows where key is null
UPDATE notification_templates SET key = name WHERE key IS NULL;

-- Update seed templates to use locale='fa' for Persian-language templates
UPDATE notification_templates SET locale = 'fa' WHERE name = 'otp_verification' AND type = 'sms';
