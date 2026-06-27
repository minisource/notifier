-- Migration: add_notification_locale
-- Version: 10
-- Description: Add locale column to notifications table for i18n support

ALTER TABLE notifications ADD COLUMN IF NOT EXISTS locale VARCHAR(10) NOT NULL DEFAULT 'en';
CREATE INDEX IF NOT EXISTS idx_notifications_locale ON notifications(locale);
