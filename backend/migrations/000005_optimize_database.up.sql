-- Database Optimization - Add advanced indexes and performance improvements
-- Migration: 000005_optimize_database

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ==========================================
-- NOTIFICATIONS TABLE OPTIMIZATIONS
-- ==========================================

-- Composite index for user notification queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_tenant_user_status 
ON notifications(tenant_id, user_id, status, created_at DESC);

-- Index for pending notifications
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_pending 
ON notifications(tenant_id, status, scheduled_at) 
WHERE status IN ('PENDING', 'SCHEDULED');

-- Index for notification type filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_tenant_type 
ON notifications(tenant_id, type, created_at DESC);

-- Index for retry logic
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_retry 
ON notifications(status, retry_count, next_retry_at) 
WHERE status = 'FAILED' AND retry_count < 3;

-- Trigram index for content search
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_content_trgm 
ON notifications USING gin(content gin_trgm_ops);

-- ==========================================
-- NOTIFICATION TEMPLATES OPTIMIZATIONS
-- ==========================================

-- Unique index for template code per tenant
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_templates_tenant_code 
ON notification_templates(tenant_id, code) 
WHERE is_active = true;

-- Index for template type lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_templates_tenant_type_channel 
ON notification_templates(tenant_id, type, channel) 
WHERE is_active = true;

-- Trigram indexes for template search
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_templates_name_trgm 
ON notification_templates USING gin(name gin_trgm_ops);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_templates_subject_trgm 
ON notification_templates USING gin(subject gin_trgm_ops);

-- ==========================================
-- NOTIFICATION PREFERENCES OPTIMIZATIONS
-- ==========================================

-- Composite index for preference lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_preferences_tenant_user_type 
ON notification_preferences(tenant_id, user_id, notification_type);

-- Index for enabled preferences
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_preferences_enabled 
ON notification_preferences(tenant_id, user_id) 
WHERE is_enabled = true;

-- ==========================================
-- NOTIFICATION LOGS OPTIMIZATIONS
-- ==========================================

-- Index for notification log queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notification_logs_notification 
ON notification_logs(notification_id, created_at DESC);

-- Index for error tracking
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notification_logs_errors 
ON notification_logs(created_at DESC) 
WHERE status = 'ERROR';

-- ==========================================
-- SMS TEMPLATES OPTIMIZATIONS
-- ==========================================

-- Unique index for SMS template code
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_sms_templates_tenant_code 
ON sms_templates(tenant_id, code) 
WHERE is_active = true;

-- Index for SMS template lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sms_templates_tenant_active 
ON sms_templates(tenant_id, is_active);

-- ==========================================
-- SETTINGS TABLE OPTIMIZATIONS
-- ==========================================

-- Unique index for settings key per tenant
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_notifier_settings_tenant_key 
ON settings(tenant_id, key);

-- ==========================================
-- AUDIT LOGS OPTIMIZATIONS
-- ==========================================

-- Partitioned index for time-based queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_time_action 
ON audit_logs(created_at DESC, action, tenant_id);

-- Index for entity tracking
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_entity 
ON audit_logs(tenant_id, entity_type, entity_id, created_at DESC);

-- Index for user activity tracking
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_activity 
ON audit_logs(tenant_id, user_id, created_at DESC) 
WHERE user_id IS NOT NULL;

-- ==========================================
-- STATISTICS AND QUERY PLANNING
-- ==========================================

-- Update statistics for better query plans
ANALYZE notifications;
ANALYZE notification_templates;
ANALYZE notification_preferences;
ANALYZE notification_logs;
ANALYZE sms_templates;
ANALYZE settings;
ANALYZE audit_logs;

-- ==========================================
-- VACUUM AND MAINTENANCE
-- ==========================================

-- Configure autovacuum for high-activity tables
ALTER TABLE notifications SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
);

ALTER TABLE notification_logs SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
);

ALTER TABLE audit_logs SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
);

-- ==========================================
-- COMMENTS FOR DOCUMENTATION
-- ==========================================

COMMENT ON INDEX idx_notifications_tenant_user_status IS 'Optimizes user notification queries';
COMMENT ON INDEX idx_notifications_pending IS 'Supports notification queue processing';
COMMENT ON INDEX idx_notifications_retry IS 'Optimizes failed notification retry logic';
COMMENT ON INDEX idx_templates_tenant_code IS 'Ensures unique template codes per tenant';
COMMENT ON INDEX idx_audit_logs_time_action IS 'Optimizes audit log queries by time and action';
