package clock

import (
	"context"
	"time"
)

type SegmentScheduler struct {
	clk           ProcessClock
	ventStepsDone int
}

func NewSegmentScheduler(clk ProcessClock) *SegmentScheduler {
	return &SegmentScheduler{clk: clk}
}

func (s *SegmentScheduler) VentStepsDone() int { return s.ventStepsDone }

type VentPlan struct {
	VentSteps int
}

func (s *SegmentScheduler) InstallVentPlan(settings VentPlan, planID string) {
	_ = s.InstallVentPlanCtx(context.Background(), settings, planID)
}

func (s *SegmentScheduler) InstallVentPlanCtx(ctx context.Context, settings VentPlan, planID string) error {
	steps := settings.VentSteps
	if steps <= 0 {
		steps = 60
	}
	for i := 0; i < steps; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.ventStepsDone = i + 1
		s.clk.Step()
		time.Sleep(2 * time.Millisecond)
	}
	return nil
}
