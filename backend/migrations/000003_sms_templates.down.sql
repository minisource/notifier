-- Migration: sms_templates (down)
-- Version: 2
-- Description: Drop SMS provider-specific template mappings

DROP TABLE IF EXISTS sms_templates;
