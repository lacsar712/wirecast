package moldloop

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/wirecast/internal/airflow"
	"github.com/lacsar712/wirecast/internal/clock"
	"github.com/lacsar712/wirecast/internal/config"
	"github.com/lacsar712/wirecast/internal/model"
	"github.com/lacsar712/wirecast/internal/shelltemp"
	"github.com/lacsar712/wirecast/internal/store"
)

func TestZoneTable(t *testing.T) {
	cfg := config.Default()
	z, err := NewZoneTable("tower-a1", 4, "plenum-main", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(z.Zones()) != 4 || z.EnabledCount() != 4 {
		t.Fatal("zone count")
	}
	zone := z.Zones()[0].Zone
	z.UpdateMoisture(zone, 16.5)
	z.UpdateFlow(zone, 800)
}

func TestCasterPlant(t *testing.T) {
	cfg := config.Default()
	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewProcessClock(start, time.Millisecond)
	mem := store.NewMemory()
	plant := NewCasterPlant(cfg, clk, mem)
	plant.Plenums().Add(airflow.NewPlenum("plenum-main", 1000))
	plant.BindAirflow("plenum-main", model.AirflowSetpoint{CubicMetersPerHour: 850, TolerancePct: 5})
	plant.Fans().Add(airflow.NewFan("fan-1"))
	if err := plant.PrimePlenum(context.Background(), "plenum-main"); err != nil {
		t.Fatal(err)
	}
	plant.ObserveFlow("plenum-main", 850)
	if err := plant.ValidateFlows(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestCoordinator(t *testing.T) {
	cfg := config.Default()
	clk := clock.NewProcessClock(time.Now(), time.Millisecond)
	bank := airflow.NewFanBank()
	bank.Add(airflow.NewFan("fan-1"))
	coord := NewCoordinator(cfg, clk, bank)
	if err := coord.Start(context.Background(), "fan-1"); err != nil {
		t.Fatal(err)
	}
	if err := coord.RampTo(context.Background(), "fan-1", 80, 4); err != nil {
		t.Fatal(err)
	}
}

func TestMoistureGradient(t *testing.T) {
	cfg := config.Default()
	clk := clock.NewProcessClock(time.Now(), time.Millisecond)
	plant := NewCasterPlant(cfg, clk, store.NewMemory())
	z1, _ := model.ParseZoneID("tower-a1", 0)
	z2, _ := model.ParseZoneID("tower-a1", 1)
	plant.RegisterSensor(shelltemp.NewSensor("s1", z1))
	plant.RegisterSensor(shelltemp.NewSensor("s2", z2))
	if err := plant.InitProfile([]model.ZoneID{z1, z2}, []float64{14, 14}); err != nil {
		t.Fatal(err)
	}
	plant.ObserveMoisture(z1, 18)
	plant.ObserveMoisture(z2, 18.5)
	if err := plant.ValidateGradient(); err != nil {
		t.Fatal(err)
	}
}

func TestZoneController(t *testing.T) {
	zc := NewZoneController("zone-1", "plenum-main")
	if !zc.Enabled() {
		t.Fatal("expected enabled")
	}
	zc.Disable()
	if zc.Enabled() {
		t.Fatal("expected disabled")
	}
}
