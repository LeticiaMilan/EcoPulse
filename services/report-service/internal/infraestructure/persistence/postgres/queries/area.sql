-- name: CreateArea :one
INSERT INTO affected_areas
    (id, type, status, area_m2, center_lat, center_lng, polygon, updated_at, votes_still_happening, votes_resolved, votes_needed)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (id) DO UPDATE SET
    type = EXCLUDED.type,
    status = EXCLUDED.status,
    area_m2 = EXCLUDED.area_m2,
    center_lat = EXCLUDED.center_lat,
    center_lng = EXCLUDED.center_lng,
    polygon = EXCLUDED.polygon,
    votes_still_happening = EXCLUDED.votes_still_happening,
    votes_resolved = EXCLUDED.votes_resolved,
    votes_needed = EXCLUDED.votes_needed,
    updated_at = NOW()
RETURNING *;