package domain

import (
	"time"

	"github.com/google/uuid"
)

type InboxStatus string

const (
	InboxStatusPending  InboxStatus = "PENDING"
	InboxStatusRejected InboxStatus = "REJECTED"
	InboxStatusAccepted InboxStatus = "PROCESSED"
)

type InboxEventType string

const (
	InboxGeneratedReportEvent   InboxEventType = "EVENTS.REPORT.GENERATED"
	InboxValidatedReportEvent   InboxEventType = "EVENTS.VALIDATIONRESPONSE.REPORT"
	InboxAreaDetectedEvent      InboxEventType = "EVENTS.AFFECTED_AREA.DETECTED"
	InboxAreaUpdatedEvent       InboxEventType = "EVENTS.AFFECTED_AREA.UPDATED"
	InboxAreaResolvedEvent      InboxEventType = "EVENTS.AFFECTED_AREA.RESOLVED"
	InboxAreaExpiredEvent       InboxEventType = "EVENTS.AFFECTED_AREA.EXPIRED"
)

// Eventos do serviço de área afetada. Todos trazem o snapshot completo da área:
// DETECTED/UPDATED ligam os relatos à área; RESOLVED/EXPIRED só atualizam o status dela.
var InboxAreaEvents = []InboxEventType{
	InboxAreaDetectedEvent,
	InboxAreaUpdatedEvent,
	InboxAreaResolvedEvent,
	InboxAreaExpiredEvent,
}

func IsAreaEvent(eventType string) bool {
	for _, areaEvent := range InboxAreaEvents {
		if string(areaEvent) == eventType {
			return true
		}
	}
	return false
}

type Inbox struct {
	ID           uuid.UUID
	EventType    InboxEventType
	Status       InboxStatus
	Content      []byte
	Retries      int32
	Received_at  time.Time
	Processed_at time.Time
	Next_attempt time.Time
}
