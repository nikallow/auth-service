-- +goose Up
CREATE TABLE sessions
(
    id           UUID PRIMARY KEY     DEFAULT uuidv7(),
    user_id      UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    refresh_hash TEXT        NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_agent   TEXT        NULL,
    ip           TEXT        NULL
);

CREATE INDEX idx_sessions_user_id ON sessions (user_id);
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);

-- +goose Down
DROP TABLE IF EXISTS sessions;
