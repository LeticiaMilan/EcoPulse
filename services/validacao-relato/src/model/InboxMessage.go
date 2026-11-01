package model

import (
	"time"
)

type InboxMessage struct {
	ID          string     `db:"id"`
	EventType   string     `db:"event_type"`
	Status      string     `db:"status"`
	Content     []byte     `db:"content"`
	Retries     int        `db:"retries"`
	ReceivedAt  time.Time  `db:"received_at"`
	NextAttempt time.Time  `db:"next_attempt"`
	ProcessedAt *time.Time `db:"processed_at"`
}
