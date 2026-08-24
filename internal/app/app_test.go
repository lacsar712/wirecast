package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/wirecast/internal/config"
	"github.com/lacsar712/wirecast/internal/model"
)

func TestRunOnce(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if err := a.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	state := a.towerFSM.State()
	if state != model.TowerCool && state != model.TowerIdle && state != model.TowerDrying {
		t.Fatalf("unexpected state %s", state)
	}
}

func TestApplyScheduleSnapshot(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	now := a.clk.Now()
	a.sched.Save(model.DryingSchedule{ID: "sch1", Entries: []model.DryingScheduleEntry{{
		Start: now.Add(-time.Hour), End: now.Add(time.Hour), Plenum: "plenum-main",
		Setpoint: model.AirflowSetpoint{CubicMetersPerHour: 800, TolerancePct: 5},
		TargetMoistPct: 14, EqualizeMinutes: 1,
	}}})
	if err := a.ApplyScheduleSnapshot(context.Background(), "sch1"); err != nil {
		t.Fatal(err)
	}
}

func TestStatusLine(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	line := a.StatusLine()
	if line == "" || !contains(line, "tower=") {
		t.Fatal(line)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
