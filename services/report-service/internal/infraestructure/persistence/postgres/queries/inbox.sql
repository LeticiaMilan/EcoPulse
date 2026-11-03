-- name: CreateInbox :one
INSERT INTO inbox
    (id, event_type, status, content, retries, received_at, processed_at)
VALUES
    ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: PoolUnprocessedInboxMessage :one
UPDATE inbox SET status = 'PROCESSED' WHERE id = (SELECT id FROM inbox WHERE status = 'PENDING' and retries < 5 and (next_attempt <= NOW() or next_attempt is null) LIMIT 1 FOR UPDATE SKIP LOCKED) RETURNING *;

-- name: UpdateInboxRetry :one
UPDATE inbox SET next_attempt = $1, retries = $2 where id = $3 returning *;