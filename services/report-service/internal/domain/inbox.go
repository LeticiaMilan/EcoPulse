package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type InboxStatus string

var (
	ErrInboxProcessingFailed = errors.New("inbox processing failed")
)

const (
	InboxStatusPending  InboxStatus = "PENDING"
	InboxStatusRejected InboxStatus = "REJECTED"
	InboxStatusAccepted InboxStatus = "PROCESSED"
)

type InboxEventType string

const (
	GeneratedReportEvent   InboxEventType = "EVENTS.GENERATED.REPORT"
	RejectedReportEvent    InboxEventType = "EVENTS.REJECTED.REPORT"
	ValidatedReportEvent   InboxEventType = "EVENTS.VALIDATED.REPORT"
	AreaDefinedReportEvent InboxEventType = "EVENTS.AREADEFINED.REPORT"
)

type inboxPayload struct {
}

type Inbox struct {
	ID           uuid.UUID
	EventType    InboxEventType
	Status       InboxStatus
	Content      []byte
	Retries      int32
	Received_at  time.Time
	Processed_at time.Time
}
