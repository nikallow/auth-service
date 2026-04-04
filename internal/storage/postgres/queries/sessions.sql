-- name: CreateSession :one
INSERT INTO sessions (user_id,
                      refresh_hash,
                      expires_at,
                      user_agent,
                      ip)
VALUES ($1,
        $2,
        $3,
        $4,
        $5)
RETURNING *;

-- name: GetSessionByID :one
SELECT *
FROM sessions
WHERE id = $1
LIMIT 1;

-- name: GetSessionByRefreshHash :one
SELECT *
FROM sessions
WHERE refresh_hash = $1
LIMIT 1;

-- name: GetActiveSessionByRefreshHash :one
SELECT *
FROM sessions
WHERE refresh_hash = $1
  AND revoked_at IS NULL
LIMIT 1;

-- name: RotateSessionRefreshHash :one
UPDATE sessions
SET refresh_hash = $2,
    expires_at   = $3
WHERE id = $1
  AND revoked_at IS NULL RETURNING *;

-- name: RevokeSession :one
UPDATE sessions
SET revoked_at = now()
WHERE id = $1
  AND revoked_at IS NULL
RETURNING *;

-- name: RevokeAllUserSessions :exec
UPDATE sessions
SET revoked_at = now()
WHERE user_id = $1
  AND revoked_at IS NULL;
