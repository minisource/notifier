# Database — Minisource Notifier

## Database Type

PostgreSQL 16+

## Schema

The notifier service uses GORM for ORM. Tables are managed via GORM AutoMigrate and/or SQL migrations.

## Tables

### notifications

Main table for all notification records.

| Column | Type | Description | Index |
|--------|------|-------------|-------|
| `id` | UUID (PK) | Auto-generated | Primary |
| `tenant_id` | UUID (nullable) | Multi-tenant isolation | Yes |
| `user_id` | UUID (not null) | Target user | Yes |
| `type` | VARCHAR(20) | sms/email/push/in_app | Yes |
| `status` | VARCHAR(20) | pending/queued/sending/sent/delivered/failed/retrying/dead/canceled/digested | Yes |
| `priority` | VARCHAR(20) | low/normal/high/urgent | No |
| `recipient_email` | VARCHAR(255) | Email recipient | Yes |
| `recipient_phone` | VARCHAR(20) | SMS recipient | Yes |
| `recipient_id` | VARCHAR(255) | Push recipient ID | Yes |
| `subject` | VARCHAR(500) | Notification subject | No |
| `body` | TEXT | Notification body | No |
| `metadata` | JSONB | Custom metadata | No |
| `template_id` | UUID (nullable) | Linked template | Yes |
| `template_key` | VARCHAR(255) | Template lookup key | Yes |
| `retry_count` | INTEGER | Current retry attempt | No |
| `max_retries` | INTEGER | Max allowed retries | No |
| `next_retry_at` | TIMESTAMP | Scheduled retry time | Yes |
| `error_message` | TEXT | Last error message | No |
| `locale` | VARCHAR(10) | Language locale | Yes |
| `idempotency_key` | VARCHAR(255) | Dedup key | Unique |
| `provider` | VARCHAR(100) | Provider used | No |
| `provider_msg_id` | VARCHAR(255) | Provider message ID | Yes |
| `scheduled_at` | TIMESTAMP | Scheduled delivery | Yes |
| `sent_at` | TIMESTAMP | Actual send time | Yes |
| `delivered_at` | TIMESTAMP | Delivery confirmation | No |
| `failed_at` | TIMESTAMP | Failure time | No |
| `seen_at` | TIMESTAMP | Displayed to user | No |
| `read_at` | TIMESTAMP | Read by user | No |
| `clicked_at` | TIMESTAMP | Clicked by user | No |
| `cancelled_at` | TIMESTAMP | Cancellation time | No |
| `created_at` | TIMESTAMP | Creation time | No |
| `updated_at` | TIMESTAMP | Last update | No |
| `deleted_at` | TIMESTAMP (soft) | Soft delete | Yes |

### notification_logs

Detailed logs of notification operations (attempt history).

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID (PK) | Auto-generated |
| `tenant_id` | UUID (nullable) | Tenant isolation |
| `notification_id` | UUID (FK) | Parent notification |
| `action` | VARCHAR(50) | created/sending/sent/failed/retrying |
| `status` | VARCHAR(20) | Current status at log time |
| `message` | TEXT | Log message |
| `error_details` | TEXT | Error details if failed |
| `provider_response` | JSONB | Provider response (sanitized) |
| `processing_time_ms` | INTEGER | Processing duration |
| `created_at` | TIMESTAMP | Log timestamp |

### notification_templates

Reusable notification templates.

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID (PK) | Auto-generated |
| `key` | VARCHAR(255) | Unique template key |
| `name` | VARCHAR(255) | Template name |
| `type` | VARCHAR(20) | sms/email/push/in_app |
| `locale` | VARCHAR(10) | Language |
| `subject` | VARCHAR(500) | Subject template |
| `body` | TEXT | Body template |
| `description` | TEXT | Template description |
| `is_active` | BOOLEAN | Active status |
| `created_at` | TIMESTAMP | Creation time |

### notification_preferences

Per-user notification preferences.

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID (PK) | Auto-generated |
| `user_id` | UUID (FK) | User |
| `type` | VARCHAR(20) | Notification channel |
| `is_enabled` | BOOLEAN | Channel enabled |
| `allow_instant` | BOOLEAN | Allow immediate send |
| `allow_digest` | BOOLEAN | Allow digest |
| `digest_frequency` | VARCHAR(20) | Frequency setting |
| `quiet_hours` | JSONB | Quiet hours config |
| `category_settings` | JSONB | Per-category settings |

### reminders

Scheduled notification reminders.

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID (PK) | Auto-generated |
| `user_id` | UUID (FK) | Target user |
| `type` | VARCHAR(20) | Notification channel |
| `recipient` | VARCHAR(255) | Recipient address |
| `subject` | VARCHAR(500) | Reminder subject |
| `body` | TEXT | Reminder body |
| `scheduled_at` | TIMESTAMP | When to send |
| `status` | VARCHAR(20) | pending/sent/cancelled |
| `notification_id` | UUID (nullable) | Resulting notification |

### settings

Dynamic configuration store.

| Column | Type | Description |
|--------|------|-------------|
| `key` | VARCHAR(255) (PK) | Setting key |
| `value` | TEXT | Setting value |
| `category` | VARCHAR(50) | Setting category |

### sms_templates

Provider-specific SMS template mappings.

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID (PK) | Auto-generated |
| `key` | VARCHAR(255) | Template key |
| `provider` | VARCHAR(255) | Provider name |
| `locale` | VARCHAR(10) | Language |
| `provider_template` | VARCHAR(255) | Provider-side template |
| `message_template` | TEXT | Fallback message |
| `token_mapping` | JSONB | Token mapping |
| `is_active` | BOOLEAN | Active status |

## Migration Strategy

The project uses **SQL migrations via `golang-migrate/migrate`** as the source of truth for production schema.

GORM AutoMigrate (`DB_AUTO_MIGRATE=true`) is for **local development only**.

### Commands

```bash
# Apply pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Rollback all
make migrate-down-all

# Check status
make migrate-status

# Create new migration
make migrate-create name=add_webhooks

# Force version (fix dirty state)
make migrate-force version=42
```

## Production Considerations

- Enable `POSTGRES_SSLMODE=require` in production
- Use a dedicated, least-privilege database user
- Enable automated backups before deployment
- Test migrations against a staging DB before production
- Monitor connection pool usage
- Indexes are already optimized for common queries
