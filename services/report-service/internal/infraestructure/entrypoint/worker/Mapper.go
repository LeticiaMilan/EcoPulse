package worker

import (
	"encoding/json"
	"report-service/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type EventEnvelope struct {
	ID          uuid.UUID       `json:"id"`
	Source      string          `json:"source"`
	ContentType string          `json:"contentType"`
	Data        json.RawMessage `json:"data"`
	RoutingKey  string          `json:"-"`
	Tx          pgx.Tx          `json:"-"`
}

func MapInboxEventToApplication(EventEnvelope EventEnvelope) *domain.Inbox {
	inboxRetorno := &domain.Inbox{
		ID:           EventEnvelope.ID,
		EventType:    domain.InboxEventType(EventEnvelope.RoutingKey),
		Status:       domain.InboxStatusPending,
		Content:      []byte(EventEnvelope.Data),
		Retries:      0,
		Received_at:  time.Now(),
		Processed_at: time.Time{},
	}

	return inboxRetorno
}

func mapOutboxEventToEnvelope(outbox *domain.Outbox) EventEnvelope {
	eventEnvelope := EventEnvelope{
		ID:          outbox.ID,
		Source:      "SERVICE.REPORT",
		RoutingKey:  string(outbox.EventType),
		ContentType: "application/json",
		Data:        outbox.Content,
	}
	return eventEnvelope
}
