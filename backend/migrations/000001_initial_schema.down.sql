-- Rollback: initial_schema
-- Version: 1

-- Drop all tables in reverse order (respecting foreign keys)

DROP TABLE IF EXISTS settings CASCADE;
DROP TABLE IF EXISTS notification_preferences CASCADE;
DROP TABLE IF EXISTS notification_logs CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS notification_templates CASCADE;

-- Note: We don't drop extensions as they might be used by other databases
-- DROP EXTENSION IF EXISTS "uuid-ossp";
-- DROP EXTENSION IF EXISTS "pgcrypto";
