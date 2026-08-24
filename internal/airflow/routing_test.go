package airflow

import (
	"context"
	"testing"

	"github.com/lacsar712/wirecast/internal/model"
)

func TestStagePlanExecute(t *testing.T) {
	bank := NewFanBank()
	bank.Add(NewFan("fan-1"))
	stager := NewStager(bank, nil)
	plan := BuildStagingPlan(4, 850, []model.FanID{"fan-1"})
	if plan.StageCount() != 3 {
		t.Fatalf("expected 3 stages, got %d", plan.StageCount())
	}
	if err := stager.Execute(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	if stager.RunningCount() != 1 {
		t.Fatal("expected running fan")
	}
}

func TestRoutePlannerAllocate(t *testing.T) {
	plenums := NewPlenumTable()
	plenums.Add(NewPlenum("plenum-main", 1000))
	router := NewRouter([]model.PlenumRoute{{
		From: "plenum-main", To: "plenum-alt", Damper: "d1", Priority: 5,
	}})
	planner := NewRoutePlanner(router, plenums)
	zones := []model.ZoneID{"z1", "z2", "z3", "z4"}
	alloc := planner.AllocateFlow("plenum-main", zones, 800)
	if len(alloc) != 4 {
		t.Fatal("allocation size")
	}
	if alloc["z1"] != 200 {
		t.Fatalf("expected 200 cmh per zone, got %v", alloc["z1"])
	}
}

func TestRoutePlannerValidate(t *testing.T) {
	plenums := NewPlenumTable()
	p := NewPlenum("plenum-main", 500)
	p.BindSetpoint(model.AirflowSetpoint{CubicMetersPerHour: 900, TolerancePct: 5})
	plenums.Add(p)
	planner := NewRoutePlanner(NewRouter(nil), plenums)
	if err := planner.ValidateRouting(context.Background()); err == nil {
		t.Fatal("expected capacity violation")
	}
}

func TestZoneFlowTable(t *testing.T) {
	tbl := NewZoneFlowTable()
	tbl.Set("z1", 200)
	tbl.Set("z2", 300)
	if tbl.Total() != 500 {
		t.Fatal("total mismatch")
	}
}
