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
	InboxAreaDefinedReportEvent InboxEventType = "EVENTS.AFFECTED_AREA.RESOLVED"
)

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
