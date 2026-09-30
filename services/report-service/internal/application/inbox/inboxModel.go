package inbox

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type EventEnvelope struct {
	ID          uuid.UUID `json:"id"`
	Source      string    `json:"source"`
	Type        string    `json:"type"`
	Subject     string    `json:"subject"`
	Time        time.Time `json:"time"`
	ContentType string    `json:"contentType"`
	Data        any       `json:"data"`
}
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

type Inbox struct {
	ID           uuid.UUID
	EventType    string
	Status       InboxStatus
	Content      []byte
	Retries      int8
	Received_at  time.Time
	Processed_at time.Time
}
