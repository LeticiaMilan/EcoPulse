package mappers

import (
	"report-service/internal/application/inbox"
	"report-service/internal/infraestructure/persistence/postgres/db"
)

func MapInboxApplicationToModel(inboxApplication *inbox.Inbox) *db.Inbox {
	return &db.Inbox{}
}

func MapInboxEventToApplication(EventEnvelope inbox.EventEnvelope) *inbox.Inbox {
	return &inbox.Inbox{}
}

func MapInboxModelToApplication(inboxModel *db.Inbox) *inbox.Inbox {
	return &inbox.Inbox{}
}
