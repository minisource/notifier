-- Migration: add_bale_safir_setting
-- Version: 11
-- Description: Add Bale Safir SMS provider configuration setting

INSERT INTO settings (key, value, category, description, is_active)
VALUES (
    'sms.providers.bale_safir',
    '{
        "accessKey": "YOUR_ACCESS_KEY",
        "botId": 0,
        "clientId": "YOUR_CLIENT_ID",
        "clientSecret": "YOUR_CLIENT_SECRET",
        "mode": "message"
    }',
    'sms',
    'Bale Safir SMS provider configuration. Supports "message" (V3) and "otp" (V2) modes.',
    true
)
ON CONFLICT (key) DO NOTHING;
