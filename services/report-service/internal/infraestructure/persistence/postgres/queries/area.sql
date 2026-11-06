-- name: CreateArea :one
INSERT INTO affected_areas
    (id, type, status, area_m2, center_lat, center_lng, polygon, updated_at)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;