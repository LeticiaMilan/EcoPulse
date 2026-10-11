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
type Outbox struct {
	ID           uuid.UUID
	EventType    OutboxEventType
	Status       OutboxStatus
	Content      []byte
	Retries      int32
	Created_at   time.Time
	Processed_at time.Time
}
