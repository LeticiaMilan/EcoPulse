package BrasilNtp

import (
	"time"

	"github.com/beevik/ntp"
)

type BrasilNTPClock struct {
	offset time.Duration
}

func NewBrasilNtpClock() (*BrasilNTPClock, error) {
	response, err := ntp.Query("a.ntp.br")

	if err != nil {
		return &BrasilNTPClock{}, err
	}

	return &BrasilNTPClock{
		offset: response.ClockOffset,
	}, nil
}

func (c *BrasilNTPClock) Now() time.Time {
	return time.Now().Add(c.offset)
}
