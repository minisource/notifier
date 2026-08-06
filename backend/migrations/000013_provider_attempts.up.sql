-- 000013_provider_attempts.up.sql
-- Durable, searchable, sanitized provider request lifecycle logging.
-- One row per outbound provider request; retries/fallbacks are distinct rows
-- linked via attempt_number / fallback_sequence / parent_attempt_id.

CREATE TABLE IF NOT EXISTS notification_provider_attempts (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                   UUID,
    notification_id             UUID NOT NULL,
    provider_account_id         UUID,
    parent_attempt_id           UUID,

    channel                     VARCHAR(20) NOT NULL,
    provider                    VARCHAR(100) NOT NULL,
    attempt_number              INTEGER NOT NULL DEFAULT 1,
    fallback_sequence           INTEGER NOT NULL DEFAULT 0,

    status                      VARCHAR(30) NOT NULL DEFAULT 'queued',
    provider_status             VARCHAR(100),
    provider_message_id         VARCHAR(255),
    provider_error_code         VARCHAR(100),
    normalized_error_kind       VARCHAR(50),
    normalized_error_code       VARCHAR(100),
    normalized_error_message    TEXT,
    retryable                   BOOLEAN NOT NULL DEFAULT FALSE,

    request_method              VARCHAR(10),
    request_url_sanitized       TEXT,
    request_headers_sanitized   JSONB NOT NULL DEFAULT '{}',
    request_body_sanitized      TEXT,
    request_size_bytes          INTEGER,

    response_status_code        INTEGER,
    response_headers_sanitized  JSONB NOT NULL DEFAULT '{}',
    response_body_sanitized     TEXT,
    response_size_bytes         INTEGER,

    body_truncated              BOOLEAN NOT NULL DEFAULT FALSE,
    original_size_bytes         INTEGER,
    captured_size_bytes         INTEGER,
    content_hash                VARCHAR(128),
    body_preview                TEXT,
    recipient_masked            VARCHAR(255),

    queued_at                   TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at                  TIMESTAMPTZ,
    completed_at                TIMESTAMPTZ,
    duration_ms                 BIGINT,
    timeout_ms                  INTEGER,

    request_id                  VARCHAR(64),
    correlation_id              VARCHAR(64),
    trace_id                    VARCHAR(64),
    span_id                     VARCHAR(64),

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS notification_provider_attempt_events (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id                  UUID NOT NULL,
    event_type                  VARCHAR(50) NOT NULL,
    previous_status             VARCHAR(30),
    new_status                  VARCHAR(30),
    event_payload_sanitized     JSONB NOT NULL DEFAULT '{}',
    source                      VARCHAR(50),
    request_id                  VARCHAR(64),
    correlation_id              VARCHAR(64),
    trace_id                    VARCHAR(64),
    occurred_at                 TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for expected filters (see design-system TODO: "Potential indexes to evaluate")
CREATE INDEX IF NOT EXISTS idx_attempts_notification ON notification_provider_attempts(notification_id);
CREATE INDEX IF NOT EXISTS idx_attempts_tenant_created ON notification_provider_attempts(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_attempts_provider_created ON notification_provider_attempts(provider, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_attempts_channel_created ON notification_provider_attempts(channel, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_attempts_status_created ON notification_provider_attempts(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_attempts_provider_msg_id ON notification_provider_attempts(provider_message_id);
CREATE INDEX IF NOT EXISTS idx_attempts_correlation ON notification_provider_attempts(correlation_id);
CREATE INDEX IF NOT EXISTS idx_attempts_request_id ON notification_provider_attempts(request_id);
CREATE INDEX IF NOT EXISTS idx_attempts_duration ON notification_provider_attempts(duration_ms);
CREATE INDEX IF NOT EXISTS idx_attempts_created ON notification_provider_attempts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_attempts_parent ON notification_provider_attempts(parent_attempt_id);

CREATE INDEX IF NOT EXISTS idx_attempt_events_attempt ON notification_provider_attempt_events(attempt_id, occurred_at ASC);
CREATE INDEX IF NOT EXISTS idx_attempt_events_occurred ON notification_provider_attempt_events(occurred_at DESC);
