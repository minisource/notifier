-- Migration: add_template_locale_key (rollback)
-- Version: 6
-- Description: Remove key and locale columns from notification_templates

DROP INDEX IF EXISTS idx_template_key_tenant;

ALTER TABLE notification_templates DROP COLUMN IF EXISTS locale;
ALTER TABLE notification_templates DROP COLUMN IF EXISTS key;
