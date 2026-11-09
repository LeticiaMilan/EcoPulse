CREATE TABLE IF NOT EXISTS outbox(
    id           UUID        PRIMARY KEY,
    event_type   VARCHAR(100),
    status       VARCHAR(100),
    content      JSONB,
    retries      INT NOT NULL DEFAULT 0,
    created_at  timestamptz,
    processed_at timestamptz
);

CREATE TABLE IF NOT EXISTS inbox(
    id           UUID        PRIMARY KEY,
    event_type   VARCHAR(100),
    status       VARCHAR(100),
    content      JSONB,
    retries      INT NOT NULL DEFAULT 0,
    received_at  timestamptz,
    next_attempt timestamptz DEFAULT NOW(),
    processed_at timestamptz
);

CREATE TABLE IF NOT EXISTS affected_areas (
  id UUID PRIMARY KEY,
  type VARCHAR(50) NOT NULL,
  status VARCHAR(30) NOT NULL,
  area_m2 DOUBLE PRECISION NOT NULL,
  center_lat DOUBLE PRECISION NOT NULL,
  center_lng DOUBLE PRECISION NOT NULL,
  polygon JSONB NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Votos "a situação ainda está acontecendo?" (vindos do calculate-area-service)
ALTER TABLE affected_areas ADD COLUMN IF NOT EXISTS votes_still_happening INT NOT NULL DEFAULT 0;
ALTER TABLE affected_areas ADD COLUMN IF NOT EXISTS votes_resolved INT NOT NULL DEFAULT 0;
ALTER TABLE affected_areas ADD COLUMN IF NOT EXISTS votes_needed INT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS report(
    id           UUID        PRIMARY KEY,
    latitude     FLOAT,
    longitude    FLOAT,
    type         VARCHAR(50),
    status       VARCHAR(50),
    area UUID NULL REFERENCES affected_areas(id) ON DELETE SET NULL,
    user_id      UUID       NOT NULL,
    datetime     timestamptz
);
