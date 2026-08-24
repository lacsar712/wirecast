package shelltemp

import (
	"testing"
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

func TestSensorObserve(t *testing.T) {
	s := NewSensor("sensor-1", "zone-1")
	now := time.Now()
	r := s.Observe(18.5, now)
	if r.Pct != 18.5 || r.Zone != "zone-1" {
		t.Fatal(r)
	}
}

func TestSensorBank(t *testing.T) {
	bank := NewSensorBank()
	bank.Register(NewSensor("s1", "zone-1"))
	now := time.Now()
	r, err := bank.ObserveZone("zone-1", 16.0, now)
	if err != nil || r.Pct != 16.0 {
		t.Fatal(err, r)
	}
	if bank.ZoneCount() != 1 {
		t.Fatal("zone count")
	}
}

func TestGradientValidate(t *testing.T) {
	g := NewGradientController(2.5)
	readings := []model.MoistureReading{
		{Zone: "z1", Pct: 18},
		{Zone: "z2", Pct: 19},
	}
	if err := g.Validate(readings); err != nil {
		t.Fatal(err)
	}
	readings[1].Pct = 22
	if err := g.Validate(readings); err == nil {
		t.Fatal("expected gradient error")
	}
}

func TestProfileManager(t *testing.T) {
	pm, err := NewProfileManager(
		[]model.ZoneID{"z1", "z2"},
		[]float64{14, 14},
	)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	pm.ArmHold(now, time.Minute)
	if !pm.HoldActive(func() time.Time { return now }) {
		t.Fatal("hold should be active")
	}
	readings := []model.MoistureReading{
		{Zone: "z1", Pct: 14.1},
		{Zone: "z2", Pct: 13.9},
	}
	if !pm.AllAtTarget(readings, 0.5) {
		t.Fatal("expected at target")
	}
}

func TestSuggestAdjustments(t *testing.T) {
	g := NewGradientController(5)
	profile := model.MoistureProfile{
		Zones:   []model.ZoneID{"z1"},
		Targets: []float64{14},
	}
	adj := g.SuggestAdjustments([]model.MoistureReading{{Zone: "z1", Pct: 18}}, profile)
	if len(adj) != 1 {
		t.Fatal("expected adjustment")
	}
}
