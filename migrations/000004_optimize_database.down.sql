-- Rollback database optimizations
-- Migration: 000004_optimize_database

-- Drop audit logs indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_user_activity;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_entity;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_time_action;

-- Drop settings indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_notifier_settings_tenant_key;

-- Drop SMS templates indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_sms_templates_tenant_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_sms_templates_tenant_code;

-- Drop notification logs indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_notification_logs_errors;
DROP INDEX CONCURRENTLY IF EXISTS idx_notification_logs_notification;

-- Drop preferences indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_preferences_enabled;
DROP INDEX CONCURRENTLY IF EXISTS idx_preferences_tenant_user_type;

-- Drop templates indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_templates_subject_trgm;
DROP INDEX CONCURRENTLY IF EXISTS idx_templates_name_trgm;
DROP INDEX CONCURRENTLY IF EXISTS idx_templates_tenant_type_channel;
DROP INDEX CONCURRENTLY IF EXISTS idx_templates_tenant_code;

-- Drop notifications indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_notifications_content_trgm;
DROP INDEX CONCURRENTLY IF EXISTS idx_notifications_retry;
DROP INDEX CONCURRENTLY IF EXISTS idx_notifications_tenant_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_notifications_pending;
DROP INDEX CONCURRENTLY IF EXISTS idx_notifications_tenant_user_status;

-- Reset autovacuum settings
ALTER TABLE notifications RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);
ALTER TABLE notification_logs RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);
ALTER TABLE audit_logs RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);

-- Note: Extensions are not dropped
-- DROP EXTENSION IF EXISTS pg_trgm;
-- DROP EXTENSION IF EXISTS pg_stat_statements;
