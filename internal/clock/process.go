package clock

import (
	"context"
	"sync"
	"time"
)

type ProcessClock struct {
	mu   sync.Mutex
	now  time.Time
	tick time.Duration
}

func NewProcessClock(start time.Time, tick time.Duration) *ProcessClock {
	if tick <= 0 {
		tick = 100 * time.Millisecond
	}
	return &ProcessClock{now: start, tick: tick}
}

func (c *ProcessClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *ProcessClock) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	go func() {
		deadline := c.Now().Add(d)
		for {
			c.mu.Lock()
			cur := c.now
			c.mu.Unlock()
			if !cur.Before(deadline) {
				ch <- cur
				return
			}
			time.Sleep(c.tick / 10)
		}
	}()
	return ch
}

func (c *ProcessClock) Step() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(c.tick)
	return c.now
}

func (c *ProcessClock) Advance(d time.Duration) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	return c.now
}

func WaitUntilContext(ctx context.Context, clk Clock, target time.Time) error {
	if !clk.Now().Before(target) {
		return nil
	}
	wait := target.Sub(clk.Now())
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-clk.After(wait):
		return nil
	}
}

func WindowElapsed(clk Clock, start time.Time, duration time.Duration) bool {
	now := clk.Now()
	end := start.Add(duration)
	return !now.Before(start) && now.Before(end)
}

func WindowClosed(clk Clock, start time.Time, duration time.Duration) bool {
	return !clk.Now().Before(start.Add(duration))
}
