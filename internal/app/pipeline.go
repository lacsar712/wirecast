package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/wirecast/internal/airflow"
	"github.com/lacsar712/wirecast/internal/tundish"
	"github.com/lacsar712/wirecast/internal/model"
	"github.com/lacsar712/wirecast/internal/shelltemp"
)

func (a *App) primeAndRoute(ctx context.Context, plenum model.PlenumID) error {
	if err := a.routePlanner.ValidateRouting(ctx); err != nil {
		return err
	}
	if err := a.plant.PrimePlenum(ctx, plenum); err != nil {
		return err
	}
	route, err := a.routePlanner.SelectPath(plenum, a.cfg.DefaultAirflowCMH)
	if err != nil {
		return err
	}
	if err := a.routePlanner.BindRoute(ctx, plenum, route.To, route.Damper, model.AirflowSetpoint{
		CubicMetersPerHour: a.cfg.DefaultAirflowCMH,
		TolerancePct:       a.cfg.AirflowTolerancePct,
	}); err != nil {
		return err
	}
	zoneIDs := make([]model.ZoneID, 0, a.cfg.ZoneCount)
	for _, z := range a.zones.Zones() {
		zoneIDs = append(zoneIDs, z.Zone)
	}
	alloc := a.routePlanner.AllocateFlow(plenum, zoneIDs, a.cfg.DefaultAirflowCMH)
	a.routePlanner.ApplyAllocation(alloc, a.zoneFlows)
	for zone, cmh := range alloc {
		a.zones.UpdateFlow(zone, cmh)
	}
	return nil
}

func (a *App) stageFans(ctx context.Context) error {
	fanIDs := []model.FanID{model.FanID("fan-1")}
	if a.plant.Fans().Count() > 1 {
		fanIDs = append(fanIDs, model.FanID("fan-2"))
	}
	sel := airflow.NewStageSelector(a.cfg.ZoneCount, a.cfg.DefaultAirflowCMH, fanIDs)
	return a.stager.Execute(ctx, sel.Select())
}

func (a *App) observeZoneMoisture(ctx context.Context) error {
	for i, z := range a.zones.Zones() {
		moist := 18.0 - float64(i)*0.3
		if err := a.plant.ObserveMoisture(z.Zone, moist); err != nil {
			return err
		}
		a.zones.UpdateMoisture(z.Zone, moist)
	}
	readings := a.plant.SensorReadings()
	if err := a.plant.ValidateGradient(); err != nil {
		return err
	}
	if err := a.gradAudit.ValidateOrdered(readings); err != nil {
		return err
	}
	if err := a.tundish.ApplyCycle(ctx, readings); err != nil {
		return err
	}
	return a.tundish.ValidateGradient(readings)
}

func (a *App) runHoldWindow(ctx context.Context) error {
	now := a.clk.Now()
	a.holdEval.Arm(now, a.plant.HoldDuration(), a.cfg.TargetMoistPct)
	if err := a.towerFSM.Apply(ctx, "moisture_hold"); err != nil {
		return err
	}
	return nil
}

func (a *App) releaseHoldWindow(ctx context.Context) error {
	readings := a.plant.SensorReadings()
	if !a.holdEval.Release(a.clk.Now(), readings, 0.5) {
		return model.Wrap("app", "hold_window", model.ErrMoistureHold)
	}
	a.plant.ReleaseHold()
	return a.towerFSM.Apply(ctx, "release_hold")
}

func (a *App) sealDehumidVents() {
	a.tundish.SealAll()
}

func (a *App) routeSummaryLine(plenum model.PlenumID) string {
	summary := a.routePlanner.Summarize(plenum)
	return fmt.Sprintf("routes=%d active_vents=%d fan_load=%.0f",
		summary.RouteCount, a.tundish.Bank().ActiveCount(), a.stager.CurrentLoad())
}

func (a *App) holdWindowStatus() string {
	if a.holdMgr.AnyActive(a.clk.Now()) {
		return "hold_window=active"
	}
	if n := a.holdMgr.PendingCount(a.clk.Now()); n > 0 {
		return fmt.Sprintf("hold_window=pending(%d)", n)
	}
	return "hold_window=idle"
}

func (a *App) gradientAuditClean() bool {
	report := a.gradAudit.AuditReport(a.plant.SensorReadings())
	return report.Clean()
}

func initHoldEvaluator(profile *shelltemp.ProfileManager, windows *shelltemp.HoldWindowManager) *shelltemp.HoldWindowEvaluator {
	return shelltemp.NewHoldWindowEvaluator(profile, windows)
}

func initDehumidController(bank *tundish.VentBank, maxDelta float64, gradientFn func([]model.MoistureReading) float64) *tundish.Controller {
	c := tundish.NewController(bank, maxDelta)
	c.WithGradientFn(gradientFn)
	return c
}
