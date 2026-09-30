-- name: CreateInbox :one
INSERT INTO inbox (id, event_type, status, content, retries, received_at, processed_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

