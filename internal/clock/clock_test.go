package clock

import (
	"context"
	"testing"
	"time"
)

func TestProcessClockAdvance(t *testing.T) {
	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	clk := NewProcessClock(start, 10*time.Millisecond)
	if clk.Now() != start {
		t.Fatal("start mismatch")
	}
	adv := clk.Advance(time.Minute)
	if !adv.Equal(start.Add(time.Minute)) {
		t.Fatal("advance mismatch")
	}
}

func TestWindowElapsed(t *testing.T) {
	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	clk := NewProcessClock(start, time.Millisecond)
	if !WindowElapsed(clk, start, time.Minute) {
		t.Fatal("window should be active at start")
	}
	clk.Advance(30 * time.Second)
	if !WindowElapsed(clk, start, time.Minute) {
		t.Fatal("window should be active mid-range")
	}
	clk.Advance(31 * time.Second)
	if WindowElapsed(clk, start, time.Minute) {
		t.Fatal("window should be closed after duration")
	}
}

func TestWaitUntilContext(t *testing.T) {
	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	clk := NewProcessClock(start, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := WaitUntilContext(ctx, clk, start.Add(time.Hour)); err == nil {
		t.Fatal("expected cancel")
	}
}

func TestWallClock(t *testing.T) {
	w := NewWallClock()
	if w.Now().IsZero() {
		t.Fatal("wall clock zero")
	}
}

func TestWindowClosed(t *testing.T) {
	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	clk := NewProcessClock(start, time.Millisecond)
	if WindowClosed(clk, start, time.Minute) {
		t.Fatal("window not closed yet")
	}
	clk.Advance(time.Minute)
	if !WindowClosed(clk, start, time.Minute) {
		t.Fatal("window should be closed")
	}
}
