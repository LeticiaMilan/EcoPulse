-- Estados de relatos
CREATE TYPE estado_relato AS ENUM (
    'POSSIVEL_RELATO',
    'RELATO_VALIDADO',
    'REJEITADO'
);

-- Tabela relatos
CREATE TABLE relatos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL, -- Id do usuário recebido por fora
    tipo VARCHAR(50) NOT NULL, -- Tipo de relato
    descricao TEXT NOT NULL, 
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    votos_positivos INT DEFAULT 0,
    votos_negativos INT DEFAULT 0,
    estado estado_relato DEFAULT 'POSSIVEL_RELATO',
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    proxima_avaliacao_em TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP + INTERVAL '15 minutes', -- Aqui está o tempo padrão de 15 minutos antes de confirmar o relato
    atualizado_em TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE votos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relato_id UUID NOT NULL REFERENCES relatos(id) ON DELETE CASCADE,
    usuario_id UUID NOT NULL, -- Id do usuário que votou
    UNIQUE (relato_id, usuario_id) -- Garante que um usuário só possa votar uma vez por relato
);

-- Índice para acelerar busca de relatos por estado
CREATE INDEX idx_relatos_avaliacao ON relatos (estado, proxima_avaliacao_em);

CREATE TABLE IF NOT EXISTS outbox(
    id           UUID        PRIMARY KEY,
    event_type   VARCHAR(100),
    status       VARCHAR(100),
    content      JSONB,
    retries      INT NOT NULL DEFAULT 0,
    created_at  timestamptz,
    next_attempt timestamptz DEFAULT NOW(),
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