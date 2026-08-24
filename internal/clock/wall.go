package clock

import "time"

type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type WallClock struct{}

func NewWallClock() *WallClock { return &WallClock{} }

func (c *WallClock) Now() time.Time { return time.Now() }

func (c *WallClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
