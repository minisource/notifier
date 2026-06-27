-- Migration: add_bale_safir_setting
-- Version: 11 (down)
-- Description: Remove Bale Safir SMS provider configuration setting

DELETE FROM settings WHERE key = 'sms.providers.bale_safir';
