package mappers

import (
	"log"
	"report-service/internal/application/inbox"
	"report-service/internal/infraestructure/persistence/postgres/db"
	"time"
)

func MapInboxApplicationToModel(inboxApplication *inbox.Inbox) *db.Inbox {
	return &db.Inbox{}
}

func MapInboxEventToApplication(EventEnvelope inbox.EventEnvelope) *inbox.Inbox {
	inboxRetorno := &inbox.Inbox{
		ID:           EventEnvelope.ID,
		EventType:    EventEnvelope.Type,
		Status:       inbox.InboxStatusPending,
		Content:      []byte(EventEnvelope.Data),
		Retries:      0,
		Received_at:  time.Now(),
		Processed_at: time.Time{},
	}
	log.Print(inboxRetorno)
	return inboxRetorno
}

func MapInboxModelToApplication(inboxModel *db.Inbox) *inbox.Inbox {
	return &inbox.Inbox{}
}
