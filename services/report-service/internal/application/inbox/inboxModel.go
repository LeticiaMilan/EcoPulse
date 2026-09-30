package inbox

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

type Inbox struct {
	ID           uuid.UUID   `db:"id"`
	EventType    string      `db:"event_type"`
	Status       InboxStatus `db:"status"`
	Content      []byte      `db:"content"`
	Retries      int8        `db:"retries"`
	Received_at  time.Time   `db:"received_at"`
	Processed_at time.Time   `db:"processed_at"`
}
