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

type inboxPayload struct {
}

type inbox struct {
	ID           uuid.UUID
	EventType    string
	Status       InboxStatus
	content      inboxPayload
	retries      int8
	received_at  time.Time
	processed_at time.Time
}

func (i *inbox) Process() {

}
