-- 000014_delivery_control_idempotency.up.sql
-- Replay protection for the Global Delivery Pause high-risk controls.
-- One row per processed Pause/Resume request keyed by (actor, idempotency_key);
-- the UNIQUE constraint makes concurrent identical requests safe (a duplicate
-- insert is a replay, never a second transition). Rows are short-lived and
-- purged by the worker (bounded retention).

CREATE TABLE IF NOT EXISTS delivery_control_idempotency (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor           VARCHAR(255) NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    operation       VARCHAR(20)  NOT NULL,
    request_hash    VARCHAR(128) NOT NULL,
    state           VARCHAR(30),
    version         BIGINT,
    result_json     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_dc_idempotency_actor_key UNIQUE (actor, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_dc_idempotency_actor_key ON delivery_control_idempotency(actor, idempotency_key);
CREATE INDEX IF NOT EXISTS idx_dc_idempotency_expires ON delivery_control_idempotency(expires_at);
