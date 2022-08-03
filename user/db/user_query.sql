-- name: GetUser :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY username;

-- name: CreateUser :execresult
INSERT INTO users (
  username, email, role
) VALUES (
  ?, ?, ?
);

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;
