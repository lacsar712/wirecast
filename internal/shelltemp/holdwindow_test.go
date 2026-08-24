package shelltemp

import (
	"testing"
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

func TestHoldWindowManager(t *testing.T) {
	mgr := NewHoldWindowManager()
	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	mgr.Schedule(start, time.Minute, 14.0, "equalize")
	if !mgr.AnyActive(start) {
		t.Fatal("hold should be active")
	}
	readings := []model.MoistureReading{{Zone: "z1", Pct: 14.1}}
	if mgr.CanRelease(start.Add(20*time.Second), readings, 0.5) {
		t.Fatal("min hold not met")
	}
	if !mgr.CanRelease(start.Add(time.Minute), readings, 0.5) {
		t.Fatal("should release after min hold")
	}
}

func TestHoldWindowEvaluator(t *testing.T) {
	pm, err := NewProfileManager([]model.ZoneID{"z1"}, []float64{14})
	if err != nil {
		t.Fatal(err)
	}
	windows := NewHoldWindowManager()
	eval := NewHoldWindowEvaluator(pm, windows)
	now := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	eval.Arm(now, time.Minute, 14.0)
	if !eval.HoldActive(now) {
		t.Fatal("expected active hold")
	}
	readings := []model.MoistureReading{{Zone: "z1", Pct: 14.0}}
	if !eval.Release(now.Add(time.Minute), readings, 0.5) {
		t.Fatal("expected release")
	}
}

func TestOrderedGradientValidator(t *testing.T) {
	order := []model.ZoneID{"top", "mid", "bot"}
	v := NewOrderedGradientValidator(order, 2.5, 0.5)
	readings := []model.MoistureReading{
		{Zone: "top", Pct: 18.0},
		{Zone: "mid", Pct: 17.5},
		{Zone: "bot", Pct: 17.0},
	}
	if err := v.ValidateOrdered(readings); err != nil {
		t.Fatal(err)
	}
	report := v.AuditReport(readings)
	if !report.Clean() {
		t.Fatal("expected clean audit")
	}
	readings[2].Pct = 18.5
	if v.Passes(readings) {
		t.Fatal("expected reverse violation")
	}
}
