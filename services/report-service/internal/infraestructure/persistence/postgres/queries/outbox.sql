-- name: CreateOutbox :one
INSERT INTO outbox (id, event_type, status, content, retries, created_at, processed_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

