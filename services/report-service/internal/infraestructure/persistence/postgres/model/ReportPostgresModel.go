package model

import (
	"time"

	"github.com/google/uuid"
)

type ReportPostgresModel struct {
	ID        uuid.UUID `db:"id"`
	Latitude  float64   `db:"latitude"`
	Longitude float64   `db:"longitude"`
	Type      string    `db:"type"`
	UserID    uuid.UUID `db:"user_id"`
	Datetime  time.Time `db:"datetime"`
	Status    string    `db:"status"`
	//Area      AffectedArea
}
