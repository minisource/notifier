-- Migration: sms_templates
-- Version: 2
-- Description: Create SMS provider-specific template mappings

-- ==========================================
-- SMS TEMPLATES (Provider-specific mappings)
-- ==========================================
-- This table maps logical template keys to provider-specific templates.
-- For lookup-based providers (like Kavenegar), provider_template contains the template name.
-- For text-based providers, message_template contains the message with placeholders like {{code}}.

CREATE TABLE IF NOT EXISTS sms_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    
    -- Template identification
    key VARCHAR(100) NOT NULL,              -- Logical key: "verify", "order_placed", etc.
    provider VARCHAR(50) NOT NULL,          -- Provider name: "kavenegar", "twilio", etc.
    
    -- Provider-specific template
    provider_template VARCHAR(255),         -- For lookup providers: template name on provider panel
    message_template TEXT,                  -- For text providers: message with {{placeholders}}
    
    -- Token mapping (maps our keys to provider token names)
    -- Example: {"code": "token", "name": "token2", "amount": "token3"}
    token_mapping JSONB DEFAULT '{}',
    
    -- Metadata
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Unique constraint: one template per key+provider+tenant
CREATE UNIQUE INDEX IF NOT EXISTS idx_sms_templates_key_provider_tenant 
    ON sms_templates(tenant_id, key, provider) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sms_templates_key ON sms_templates(key);
CREATE INDEX IF NOT EXISTS idx_sms_templates_provider ON sms_templates(provider);
CREATE INDEX IF NOT EXISTS idx_sms_templates_deleted_at ON sms_templates(deleted_at);

-- ==========================================
-- SEED DEFAULT TEMPLATES
-- ==========================================
-- Insert default templates for Kavenegar (these can be overridden per tenant)

INSERT INTO sms_templates (tenant_id, key, provider, provider_template, token_mapping, description)
VALUES 
    -- Verify OTP template
    (NULL, 'verify', 'kavenegar', 'verify', '{"code": "token"}', 'OTP verification code'),
    -- You can add more templates here as needed
    (NULL, 'welcome', 'kavenegar', 'welcome', '{"name": "token"}', 'Welcome message'),
    (NULL, 'order_placed', 'kavenegar', 'orderPlaced', '{"order_id": "token", "amount": "token2"}', 'Order placed notification'),
    (NULL, 'payment_success', 'kavenegar', 'paymentSuccess', '{"amount": "token", "ref": "token2"}', 'Payment success notification')
ON CONFLICT DO NOTHING;
