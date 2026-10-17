-- name: CreateInbox :one
INSERT INTO inbox (id, event_type, status, content, retries, received_at, processed_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: PoolUnprocessedInboxMessage :one
UPDATE inbox SET status = 'PROCESSED' WHERE id = (SELECT id FROM inbox WHERE status = 'PENDING' LIMIT 1 FOR UPDATE SKIP LOCKED) RETURNING *;