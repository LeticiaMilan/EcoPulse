package events

import (
	"database/sql"
	"encoding/json"
	"time"

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
type Event struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	Version    int       `json:"version"`
	OccurredAt time.Time `json:"occurred_at"`
	Data       any       `json:"data"`
}

// Registra um evento dentro da transação recebida
func SaveOutboxEvent(
	tx *sql.Tx,
	eventType string,
	routingKey string,
	data any,
) error {

	event := Event{
		EventID:    uuid.NewString(),
		EventType:  eventType,
		Version:    1,
		OccurredAt: time.Now().UTC(),
		Data:       data,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO outbox_events
		(id, event_type, routing_key, payload)
		VALUES ($1, $2, $3, $4)
	`,
		event.EventID,
		event.EventType,
		routingKey,
		string(payload),
	)

	return err
}
