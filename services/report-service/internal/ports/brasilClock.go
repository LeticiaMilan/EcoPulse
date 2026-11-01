package ports

import "time"

type BrasilClock interface {
	Now() time.Time
}
