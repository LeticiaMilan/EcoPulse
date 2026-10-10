-- name: CreateReport :one
INSERT INTO report (id, latitude, longitude, type, status, user_id, datetime) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;
