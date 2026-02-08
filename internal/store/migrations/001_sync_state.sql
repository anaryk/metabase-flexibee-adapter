CREATE TABLE IF NOT EXISTS sync_state (
    evidence    TEXT PRIMARY KEY,
    last_update TIMESTAMPTZ,
    last_sync   TIMESTAMPTZ NOT NULL,
    row_count   BIGINT DEFAULT 0,
    status      TEXT DEFAULT 'ok',
    error_msg   TEXT
);

CREATE TABLE IF NOT EXISTS cleanup_log (
    id           BIGSERIAL PRIMARY KEY,
    evidence     TEXT NOT NULL,
    cleaned_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rows_deleted BIGINT NOT NULL,
    oldest_kept  TIMESTAMPTZ
);
