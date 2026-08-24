package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/wirecast/internal/config"
)

func TestCase(t *testing.T) {
	cfg := config.Default()
	cfg.SegmentVentSteps = 80
	a, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = a.ExecutePlan(ctx, SegmentPlan{VentSteps: cfg.SegmentVentSteps})
	}()
	time.Sleep(5 * time.Millisecond)
	cancel()
	<-done
	if a.SegmentVentStepsDone() >= cfg.SegmentVentSteps {
		t.Fatalf("segment vent plan finished after cancel: %d steps", a.SegmentVentStepsDone())
	}
}
