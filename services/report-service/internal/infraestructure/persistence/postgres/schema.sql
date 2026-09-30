CREATE TABLE IF NOT EXISTS inbox(
    id           UUID        PRIMARY KEY,
    event_type   VARCHAR(100),
    status       VARCHAR(100),
    content      JSONB,
    retries      int,
    received_at  timestamptz,
    processed_at timestamptz
);

CREATE TABLE IF NOT EXISTS report(
    id           UUID        PRIMARY KEY,
    latitude     FLOAT,
    longitude    FLOAT,
    type         VARCHAR(50),
    user_id      UUID,
    datetime     timestamptz
);