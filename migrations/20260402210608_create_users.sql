-- +goose Up
CREATE TABLE users
(
    id             UUID PRIMARY KEY     DEFAULT uuidv7(),
    email          TEXT        NOT NULL UNIQUE,
    password_hash  TEXT        NOT NULL,
    email_verified BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMP   NULL
);

-- +goose Down
DROP TABLE IF EXISTS users;
