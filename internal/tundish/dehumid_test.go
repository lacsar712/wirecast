package tundish

import (
	"context"
	"testing"

	"github.com/lacsar712/wirecast/internal/model"
)

func TestVentBankBetweenSegments(t *testing.T) {
	zones := []model.ZoneID{"tower-a1-zone-00", "tower-a1-zone-01", "tower-a1-zone-02"}
	bank, err := NewVentBank(zones, "vent")
	if err != nil {
		t.Fatal(err)
	}
	if bank.Count() != 2 {
		t.Fatalf("expected 2 vents, got %d", bank.Count())
	}
	vent, ok := bank.Between(zones[0], zones[1])
	if !ok || vent.Damper != "vent-00" {
		t.Fatal("vent lookup failed")
	}
}

func TestControllerPlanCycle(t *testing.T) {
	zones := []model.ZoneID{"z-top", "z-mid", "z-bot"}
	bank, err := NewVentBank(zones, "v")
	if err != nil {
		t.Fatal(err)
	}
	c := NewController(bank, 2.5)
	readings := []model.MoistureReading{
		{Zone: "z-top", Pct: 18.0},
		{Zone: "z-mid", Pct: 17.5},
		{Zone: "z-bot", Pct: 17.0},
	}
	actions := c.PlanCycle(readings)
	if len(actions) == 0 {
		t.Fatal("expected vent actions")
	}
	if err := c.ApplyCycle(context.Background(), readings); err != nil {
		t.Fatal(err)
	}
	if bank.ActiveCount() == 0 {
		t.Fatal("expected active vents")
	}
}

func TestControllerValidateGradient(t *testing.T) {
	bank, _ := NewVentBank([]model.ZoneID{"a", "b"}, "v")
	c := NewController(bank, 1.0)
	readings := []model.MoistureReading{{Zone: "a", Pct: 18}, {Zone: "b", Pct: 20}}
	if err := c.ValidateGradient(readings); err == nil {
		t.Fatal("expected gradient error")
	}
}
