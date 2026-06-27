-- Migration: initial_schema
-- Version: 1
-- Description: Create all initial tables for notifier service

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==========================================
-- NOTIFICATION TEMPLATES
-- ==========================================
CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    subject VARCHAR(500),
    body TEXT NOT NULL,
    description TEXT,
    variables JSONB DEFAULT '[]',
    provider VARCHAR(100),
    provider_template VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_template_name_type_tenant 
    ON notification_templates(tenant_id, name, type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notification_templates_tenant_id ON notification_templates(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notification_templates_type ON notification_templates(type);
CREATE INDEX IF NOT EXISTS idx_notification_templates_deleted_at ON notification_templates(deleted_at);

-- ==========================================
-- NOTIFICATIONS
-- ==========================================
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    user_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    
    -- Recipient information
    recipient_email VARCHAR(255),
    recipient_phone VARCHAR(20),
    recipient_id VARCHAR(255),
    
    -- Content
    subject VARCHAR(500),
    body TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    
    -- Template reference
    template_id UUID REFERENCES notification_templates(id) ON DELETE SET NULL,
    
    -- Retry information
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    
    -- Error information
    error_message TEXT,
    
    -- Provider information
    provider VARCHAR(100),
    provider_msg_id VARCHAR(255),
    
    -- Timing
    scheduled_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    read_at TIMESTAMP WITH TIME ZONE,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_notifications_tenant_id ON notifications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_email ON notifications(recipient_email);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_phone ON notifications(recipient_phone);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_id ON notifications(recipient_id);
CREATE INDEX IF NOT EXISTS idx_notifications_template_id ON notifications(template_id);
CREATE INDEX IF NOT EXISTS idx_notifications_provider_msg_id ON notifications(provider_msg_id);
CREATE INDEX IF NOT EXISTS idx_notifications_scheduled_at ON notifications(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_notifications_sent_at ON notifications(sent_at);
CREATE INDEX IF NOT EXISTS idx_notifications_next_retry_at ON notifications(next_retry_at);
CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at ON notifications(deleted_at);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_notifications_user_status ON notifications(user_id, status);
CREATE INDEX IF NOT EXISTS idx_notifications_pending_scheduled ON notifications(status, scheduled_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_notifications_retry ON notifications(status, next_retry_at) WHERE status = 'retrying';

-- ==========================================
-- NOTIFICATION LOGS
-- ==========================================
CREATE TABLE IF NOT EXISTS notification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    error_details TEXT,
    provider_response JSONB,
    processing_time_ms INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_logs_tenant_id ON notification_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notification_logs_notification_id ON notification_logs(notification_id);
CREATE INDEX IF NOT EXISTS idx_notification_logs_created_at ON notification_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_notification_logs_action ON notification_logs(action);

-- ==========================================
-- NOTIFICATION PREFERENCES
-- ==========================================
CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    user_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    allow_instant BOOLEAN NOT NULL DEFAULT TRUE,
    allow_digest BOOLEAN NOT NULL DEFAULT FALSE,
    digest_frequency VARCHAR(20) DEFAULT 'daily',
    quiet_hours JSONB,
    category_settings JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_type_tenant 
    ON notification_preferences(tenant_id, user_id, type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notification_preferences_tenant_id ON notification_preferences(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_deleted_at ON notification_preferences(deleted_at);

-- ==========================================
-- SETTINGS
-- ==========================================
CREATE TABLE IF NOT EXISTS settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(255) NOT NULL UNIQUE,
    value TEXT,
    type VARCHAR(50) DEFAULT 'string',
    category VARCHAR(100),
    description VARCHAR(500),
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key);
CREATE INDEX IF NOT EXISTS idx_settings_category ON settings(category);
CREATE INDEX IF NOT EXISTS idx_settings_deleted_at ON settings(deleted_at);

-- ==========================================
-- SEED DEFAULT DATA
-- ==========================================

-- Insert default settings
INSERT INTO settings (id, key, value, type, category, description, is_public)
VALUES 
    (gen_random_uuid(), 'email_from_address', 'noreply@example.com', 'string', 'email', 'Default from email address', FALSE),
    (gen_random_uuid(), 'email_from_name', 'Notification Service', 'string', 'email', 'Default from name for emails', FALSE),
    (gen_random_uuid(), 'sms_default_provider', 'kavenegar', 'string', 'sms', 'Default SMS provider', FALSE),
    (gen_random_uuid(), 'email_default_provider', 'smtp', 'string', 'email', 'Default email provider', FALSE),
    (gen_random_uuid(), 'max_retry_attempts', '3', 'int', 'general', 'Maximum retry attempts for failed notifications', FALSE),
    (gen_random_uuid(), 'retry_delay_seconds', '60', 'int', 'general', 'Delay between retries in seconds', FALSE),
    (gen_random_uuid(), 'batch_size', '100', 'int', 'general', 'Batch size for processing notifications', FALSE),
    (gen_random_uuid(), 'rate_limit_per_minute', '100', 'int', 'general', 'Rate limit per minute per provider', FALSE)
ON CONFLICT (key) DO NOTHING;

-- Insert default notification templates
INSERT INTO notification_templates (id, name, type, subject, body, variables, is_active)
VALUES 
    (gen_random_uuid(), 'otp_verification', 'sms', NULL, 'Your verification code is: {{code}}. Valid for {{expiryMinutes}} minutes.', '["code", "expiryMinutes"]', TRUE),
    (gen_random_uuid(), 'otp_verification', 'email', 'Verification Code', '<h1>Your Verification Code</h1><p>Your code is: <strong>{{code}}</strong></p><p>Valid for {{expiryMinutes}} minutes.</p>', '["code", "expiryMinutes"]', TRUE),
    (gen_random_uuid(), 'welcome', 'email', 'Welcome to Our Platform', '<h1>Welcome, {{userName}}!</h1><p>Thank you for joining us.</p>', '["userName"]', TRUE),
    (gen_random_uuid(), 'password_reset', 'email', 'Password Reset Request', '<h1>Password Reset</h1><p>Click <a href="{{resetLink}}">here</a> to reset your password.</p><p>This link expires in {{expiryMinutes}} minutes.</p>', '["resetLink", "expiryMinutes"]', TRUE),
    (gen_random_uuid(), 'password_reset', 'sms', NULL, 'Password reset code: {{code}}. Valid for {{expiryMinutes}} minutes.', '["code", "expiryMinutes"]', TRUE)
ON CONFLICT DO NOTHING;
