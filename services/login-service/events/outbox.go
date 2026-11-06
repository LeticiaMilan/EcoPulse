package events

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
)

const CreateOutboxTableSQL = `
CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    routing_key VARCHAR(100) NOT NULL,
    payload TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_pending
ON outbox_events (created_at)
WHERE published_at IS NULL;
`

// Estrutura padrão dos eventos
type EventEnvelope struct {
	ID          string          `json:"id"`
	Source      string          `json:"source"`
	ContentType string          `json:"contentType"`
	Data        json.RawMessage `json:"data"`
}

// Registra um evento dentro da transação recebida
func SaveOutboxEvent(
	tx *sql.Tx,
	eventType string,
	routingKey string,
	data any,
) error {
	eventID := uuid.NewString()

	// Serializa somente o payload específico do evento.
	eventData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Monta o novo envelope.
	event := EventEnvelope{
		ID:          eventID,
		Source:      "login-service",
		ContentType: "application/json",
		Data:        eventData,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO outbox_events
        (id, event_type, routing_key, payload)
        VALUES ($1, $2, $3, $4)`,
		eventID,
		eventType,
		routingKey,
		string(payload),
	)

	return err
}
