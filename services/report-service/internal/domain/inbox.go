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
	InboxRejectedReportEvent    InboxEventType = "EVENTS.REPORT.REJECTED"
	InboxValidatedReportEvent   InboxEventType = "EVENTS.REPORT.VALIDATED"
	InboxAreaDefinedReportEvent InboxEventType = "EVENTS.REPORT.AREADEFINED"
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
