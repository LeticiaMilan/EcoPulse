CREATE TABLE IF NOT EXISTS affected_areas (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    polygon JSONB NOT NULL,
    area_m2 DOUBLE PRECISION NOT NULL,
    center_lat DOUBLE PRECISION NOT NULL,
    center_lng DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    last_activity_at TIMESTAMPTZ NOT NULL,
    last_confirmation_request_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_affected_areas_type_status ON affected_areas (type, status);

CREATE TABLE IF NOT EXISTS area_reports (
    report_id UUID PRIMARY KEY,
    area_id UUID NOT NULL REFERENCES affected_areas (id),
    user_id UUID NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_area_reports_area_id ON area_reports (area_id);

-- Uma resposta por usuário em cada área (vale a mais recente)
CREATE TABLE IF NOT EXISTS area_votes (
    area_id UUID NOT NULL REFERENCES affected_areas (id),
    user_id UUID NOT NULL,
    still_happening BOOLEAN NOT NULL,
    voted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (area_id, user_id)
);

-- Status iguais aos do report-service: PENDING, PROCESSED, REJECTED
CREATE TABLE IF NOT EXISTS inbox (
    event_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    retries SMALLINT NOT NULL DEFAULT 0,
    error TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_inbox_pending ON inbox (received_at) WHERE status = 'PENDING';

CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_unpublished ON outbox (created_at) WHERE published_at IS NULL;
