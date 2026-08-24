package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/wirecast/internal/airflow"
	"github.com/lacsar712/wirecast/internal/clock"
	"github.com/lacsar712/wirecast/internal/config"
	"github.com/lacsar712/wirecast/internal/tundish"
	"github.com/lacsar712/wirecast/internal/moldloop"
	"github.com/lacsar712/wirecast/internal/fsm"
	"github.com/lacsar712/wirecast/internal/interlock"
	"github.com/lacsar712/wirecast/internal/model"
	"github.com/lacsar712/wirecast/internal/shelltemp"
	"github.com/lacsar712/wirecast/internal/store"
)

type App struct {
	cfg          config.Config
	clk          clock.Clock
	mem          *store.Memory
	sched        *store.ScheduleStore
	plant        *moldloop.CasterPlant
	towerFSM     *fsm.TowerFSM
	zones        *moldloop.ZoneTable
	lock         *interlock.DamperLock
	dampers      *airflow.DamperActuator
	router       *airflow.Router
	routePlanner *airflow.RoutePlanner
	zoneFlows    *airflow.ZoneFlowTable
	stager       *airflow.Stager
	tundish      *tundish.Controller
	holdMgr      *shelltemp.HoldWindowManager
	holdEval     *shelltemp.HoldWindowEvaluator
	gradAudit    *shelltemp.OrderedGradientValidator
	guard        *interlock.Guard
	dryFSM       *fsm.CasterFSM
	scheduler    *clock.SegmentScheduler
	avgWindow    *clock.SolidWindow
	feedLeases   *interlock.FeedLeaseRegistry
	batchMu      sync.Mutex
	batchCancels map[model.TowerID]context.CancelFunc
	dryRamp      *CastRamp
}

func New(cfg config.Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewProcessClock(start, cfg.ProcessTick())
	mem := store.NewMemory()
	towerID, err := model.ParseTowerID(cfg.TowerID)
	if err != nil {
		return nil, err
	}
	plenumID := model.PlenumID("plenum-main")
	zones, err := moldloop.NewZoneTable(towerID, cfg.ZoneCount, plenumID, cfg)
	if err != nil {
		return nil, err
	}
	plant := moldloop.NewCasterPlant(cfg, clk, mem)
	plant.Plenums().Add(airflow.NewPlenum(plenumID, 1200))
	plant.BindAirflow(plenumID, model.AirflowSetpoint{CubicMetersPerHour: cfg.DefaultAirflowCMH, TolerancePct: cfg.AirflowTolerancePct})
	fan := airflow.NewFan(model.FanID("fan-1"))
	plant.Fans().Add(fan)
	plant.Plenums().Add(airflow.NewPlenum("plenum-alt", 900))
	guardPairs := make(map[model.ZoneID]model.PlenumID)
	zoneIDs := make([]model.ZoneID, 0, cfg.ZoneCount)
	targets := make([]float64, 0, cfg.ZoneCount)
	for i, z := range zones.Zones() {
		guardPairs[z.Zone] = plenumID
		zoneIDs = append(zoneIDs, z.Zone)
		sensorID, _ := model.ParseSensorID(fmt.Sprintf("sensor-%02d", i))
		plant.RegisterSensor(shelltemp.NewSensor(sensorID, z.Zone))
		targets = append(targets, cfg.TargetMoistPct)
	}
	if err := plant.InitProfile(zoneIDs, targets); err != nil {
		return nil, err
	}
	ventBank, err := tundish.NewVentBank(zoneIDs, "vent-damper")
	if err != nil {
		return nil, err
	}
	router := airflow.NewRouter([]model.PlenumRoute{
		{From: plenumID, To: "plenum-alt", Damper: "damper-main", Priority: 10},
	})
	holdMgr := shelltemp.NewHoldWindowManager()
	a := &App{
		cfg: cfg, clk: clk, mem: mem, sched: store.NewScheduleStore(mem), plant: plant, zones: zones,
		lock: interlock.NewDamperLock(clk.Now), router: router,
		routePlanner: airflow.NewRoutePlanner(router, plant.Plenums()),
		zoneFlows:    airflow.NewZoneFlowTable(),
		stager:       airflow.NewStager(plant.Fans(), clk),
		holdMgr:      holdMgr,
		holdEval:     initHoldEvaluator(plant.Profile(), holdMgr),
		gradAudit:    shelltemp.NewOrderedGradientValidator(zoneIDs, cfg.MaxGradientDeltaPct, 0.5),
		guard:        interlock.NewGuard(guardPairs),
		batchCancels: make(map[model.TowerID]context.CancelFunc),
	}
	a.tundish = initDehumidController(ventBank, cfg.MaxGradientDeltaPct, func(readings []model.MoistureReading) float64 {
		return plant.GradientDeltaFor(readings)
	})
	a.dampers = airflow.NewDamperActuator(a.lock)
	a.dampers.Register(airflow.NewDamper("damper-main"))
	for _, vent := range ventBank.All() {
		a.dampers.Register(airflow.NewDamper(vent.Damper))
	}
	a.towerFSM = fsm.NewTowerFSM(towerID, a.onTowerTransition)
	a.feedLeases = interlock.NewFeedLeaseRegistry(a.clk.Now)
	if pc, ok := a.clk.(*clock.ProcessClock); ok {
		a.scheduler = clock.NewSegmentScheduler(*pc)
		a.avgWindow = clock.NewSolidWindow(a.clk, time.Duration(cfg.SolidWindowMinutes)*time.Minute)
	}
	a.dryFSM = fsm.NewCasterFSM(towerID, a.onDryTransition)
	fsm.RegisterDryHeatHook(a.dryFSM.Hooks())
	a.dryRamp = NewCastRamp(a.clk, cfg.ProcessTick(), cfg.CastRampSteps)
	a.persistSnapshot(towerID)
	return a, nil
}

func (a *App) onTowerTransition(ctx context.Context, tower model.TowerID, from, to model.TowerState) error {
	if to == model.TowerFault {
		return model.Wrap("app", "tower_fault", model.ErrFanFault)
	}
	return nil
}

func (a *App) onDryTransition(ctx context.Context, tower model.TowerID, from, to model.DryState) error {
	_ = ctx
	_ = tower
	_ = from
	_ = to
	return nil
}

func (a *App) persistSnapshot(id model.TowerID) {
	b := store.NewSnapshotBuilder(id).State(a.towerFSM.State())
	for _, z := range a.zones.Zones() {
		b.Zone(z)
	}
	b.Fan(model.FanID("fan-1"))
	a.mem.PutTower(b.Build(a.clk.Now()))
}

func (a *App) ApplyScheduleSnapshot(ctx context.Context, id model.ScheduleID) error {
	snap, err := a.sched.SnapshotClone(id)
	if err != nil {
		return err
	}
	now := a.clk.Now()
	entry, ok := a.sched.ActiveEntry(snap, now)
	if !ok {
		return model.Wrap("app", "schedule", model.ErrScheduleEmpty)
	}
	a.plant.BindAirflow(entry.Plenum, entry.Setpoint)
	a.plant.ArmMoistureHold(now, time.Duration(entry.EqualizeMinutes)*time.Minute, entry.TargetMoistPct)
	return nil
}

func (a *App) RunOnce(ctx context.Context) error {
	if err := a.towerFSM.Apply(ctx, "preheat"); err != nil {
		return err
	}
	plenum := model.PlenumID("plenum-main")
	if err := a.primeAndRoute(ctx, plenum); err != nil {
		return err
	}
	if err := a.dampers.Move(ctx, "damper-main", 100); err != nil {
		return err
	}
	if err := a.towerFSM.Apply(ctx, "airflow_ok"); err != nil {
		return err
	}
	if err := a.stageFans(ctx); err != nil {
		return err
	}
	a.plant.ObserveFlow(plenum, a.cfg.DefaultAirflowCMH)
	if err := a.plant.ValidateFlows(ctx); err != nil {
		return err
	}
	if err := a.observeZoneMoisture(ctx); err != nil {
		return err
	}
	a.plant.ArmMoistureHold(a.clk.Now(), time.Duration(a.cfg.EqualizeHoldMinutes)*time.Minute, a.cfg.TargetMoistPct)
	if err := a.runHoldWindow(ctx); err != nil {
		return err
	}
	if pc, ok := a.clk.(*clock.ProcessClock); ok {
		pc.Advance(time.Duration(a.cfg.EqualizeHoldMinutes)*time.Minute + time.Second)
	}
	if err := a.releaseHoldWindow(ctx); err != nil {
		return err
	}
	a.sealDehumidVents()
	for _, z := range a.zones.Zones() {
		if err := a.plant.ObserveMoisture(z.Zone, a.cfg.TargetMoistPct); err != nil {
			return err
		}
	}
	if a.plant.AtTarget(0.5) {
		if err := a.towerFSM.Apply(ctx, "target_reached"); err != nil {
			return err
		}
	}
	if a.towerFSM.State() == model.TowerCool {
		if pc, ok := a.clk.(*clock.ProcessClock); ok {
			pc.Advance(time.Duration(a.cfg.DryingRampMinutes) * time.Minute)
		}
		if err := a.towerFSM.Apply(ctx, "cool_complete"); err != nil {
			return err
		}
	}
	a.persistSnapshot(model.TowerID(a.cfg.TowerID))
	return nil
}

func (a *App) StatusLine() string {
	plenum := model.PlenumID("plenum-main")
	return fmt.Sprintf("tower=%s state=%s hold=%v zones=%d gradient=%.2f %s %s audit=%v",
		a.cfg.TowerID, a.towerFSM.State(), a.plant.HoldActive(), a.zones.EnabledCount(),
		a.plant.GradientDelta(), a.routeSummaryLine(plenum), a.holdWindowStatus(), a.gradientAuditClean())
}
