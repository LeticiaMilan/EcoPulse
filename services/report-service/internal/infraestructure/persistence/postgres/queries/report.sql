-- name: CreateReport :one
INSERT INTO report (id, latitude, longitude, type, status, user_id, datetime, area) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: ValidateReport :one
UPDATE report SET status = 'validated' WHERE id = $1 RETURNING *;

-- name: RejectReport :one
UPDATE report SET status = 'rejected' WHERE id = $1 RETURNING *;

-- name: DefineAreaReport :one
UPDATE report SET area = $1 WHERE id = $2 AND area is NULL RETURNING *;

-- name: GetReportsWithArea :many
SELECT
    r.id AS report_id,
    r.latitude,
    r.longitude,
    r.type AS report_type,
    r.status AS report_status,
    r.user_id,
    r.datetime,
    a.id AS area_id,
    a.type AS area_type,
    a.status AS area_status,
    a.area_m2,
    a.center_lat,
    a.center_lng,
    a.polygon,
    a.updated_at AS area_updated_at
FROM report r
         INNER JOIN affected_areas a ON a.id = r.area
WHERE r.datetime >= $1
ORDER BY r.datetime ASC
LIMIT $2;