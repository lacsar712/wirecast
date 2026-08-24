package app

import (
	"context"

	"github.com/lacsar712/wirecast/internal/clock"
)

// SegmentPlan describes inter-segment vent staging for drying batches.

type SegmentPlan struct {
	VentSteps int
}

func (a *App) ExecutePlan(ctx context.Context, plan SegmentPlan) error {
	if a.scheduler == nil {
		return nil
	}
	return a.scheduler.InstallVentPlanCtx(ctx, clock.VentPlan{VentSteps: plan.VentSteps}, "segment-plan")
}

func (a *App) SegmentVentStepsDone() int {
	if a.scheduler == nil {
		return 0
	}
	return a.scheduler.VentStepsDone()
}
