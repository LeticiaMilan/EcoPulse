package model

import (
	"time"
)

type OutboxMessage struct {
	ID          string     `db:"id"`
	EventType   string     `db:"event_type"`
	Status      string     `db:"status"`
	Content     []byte     `db:"content"` // JSONB pode ser mapeado como []byte ou string
	Retries     int        `db:"retries"`
	CreatedAt   time.Time  `db:"created_at"`
	NextAttempt time.Time  `db:"next_attempt"`
	ProcessedAt *time.Time `db:"processed_at"`
}
