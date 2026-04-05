-- name: CreateUser :one
INSERT INTO users (email,
                   password_hash)
VALUES ($1,
        $2)
ON CONFLICT (email) DO NOTHING
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;

-- name: GetActiveUserByID :one
SELECT *
FROM users
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetActiveUserByEmail :one
SELECT *
FROM users
WHERE email = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: MarkEmailVerified :one
UPDATE users
SET email_verified = TRUE,
    updated_at     = now()
WHERE id = $1
RETURNING *;

-- name: SoftDeleteUser :one
UPDATE users
SET deleted_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;
