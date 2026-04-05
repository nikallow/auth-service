-- +goose Up
CREATE UNIQUE INDEX idx_sessions_refresh_hash_unique ON sessions (refresh_hash);

-- +goose Down
DROP INDEX IF EXISTS idx_sessions_refresh_hash_unique;
