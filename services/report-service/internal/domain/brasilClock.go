package domain

import "time"

type BrasilClock interface {
	Now() time.Time
}
