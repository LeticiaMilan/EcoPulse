package domain

import (
	"time"

	"github.com/google/uuid"
)

type OutboxStatus string

const (
	OutboxStatusPending  OutboxStatus = "PENDING"
	OutboxStatusRejected OutboxStatus = "REJECTED"
	OutboxStatusAccepted OutboxStatus = "PROCESSED"
)

type OutboxEventType string

const (
	OutboxReportGeneratedError OutboxEventType = "EVENTS.REPORT.GENERATED_ERROR"
	OutboxReportCreated        OutboxEventType = "EVENTS.REPORT.CREATED"
)

type Outbox struct {
	ID           uuid.UUID
	EventType    OutboxEventType
	Status       OutboxStatus
	Content      []byte
	Retries      int32
	Created_at   time.Time
	Processed_at time.Time
	Next_attempt time.Time
}

func NewOutbox(eventType OutboxEventType, Content []byte) Outbox {
	return Outbox{
		ID:           uuid.New(),
		EventType:    eventType,
		Status:       OutboxStatusPending,
		Created_at:   time.Now(),
		Content:      Content,
		Retries:      0,
		Processed_at: time.Time{},
		Next_attempt: time.Time{},
	}
}
