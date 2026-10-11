package worker

import (
	"encoding/json"
	"report-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

type EventEnvelope struct {
	ID          uuid.UUID       `json:"id"`
	Source      string          `json:"source"`
	Type        string          `json:"type"`
	Time        time.Time       `json:"time"`
	ContentType string          `json:"contentType"`
	Data        json.RawMessage `json:"data"`
}

func MapInboxEventToApplication(EventEnvelope *EventEnvelope) *domain.Inbox {
	inboxRetorno := &domain.Inbox{
		ID:           EventEnvelope.ID,
		EventType:    domain.InboxEventType(EventEnvelope.Type),
		Status:       domain.InboxStatusPending,
		Content:      []byte(EventEnvelope.Data),
		Retries:      0,
		Received_at:  time.Now(),
		Processed_at: time.Time{},
	}

	return inboxRetorno
}

func mapOutboxEventToEnvelope(outbox *domain.Outbox) *EventEnvelope {
	eventEnvelope := &EventEnvelope{
		ID:          uuid.New(),
		Source:      "eco-pulse/report-service",
		Type:        string(outbox.EventType),
		Time:        time.Time{},
		ContentType: "application/json",
		Data:        outbox.Content,
	}
	return eventEnvelope
}
