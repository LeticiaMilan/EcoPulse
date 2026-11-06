-- name: CreateReport :one
INSERT INTO report (id, latitude, longitude, type, status, user_id, datetime, area) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: ValidateReport :one
UPDATE report SET status = 'validated' WHERE id = $1 RETURNING *;

-- name: RejectReport :one
UPDATE report SET status = 'rejected' WHERE id = $1 RETURNING *;

-- name: DefineAreaReport :one
UPDATE report SET area = $1 WHERE id = $2 AND area is NULL RETURNING *;