package fsm

import (
	"context"
	"testing"

	"github.com/lacsar712/wirecast/internal/model"
)

func TestTowerFSMHappyPath(t *testing.T) {
	fsm := NewTowerFSM("tower-a1", nil)
	ctx := context.Background()
	if err := fsm.Apply(ctx, "preheat"); err != nil {
		t.Fatal(err)
	}
	if fsm.State() != model.TowerPreheat {
		t.Fatal(fsm.State())
	}
	if err := fsm.Apply(ctx, "airflow_ok"); err != nil {
		t.Fatal(err)
	}
	if fsm.State() != model.TowerDrying {
		t.Fatal(fsm.State())
	}
}

func TestTowerFSMIllegal(t *testing.T) {
	fsm := NewTowerFSM("tower-a1", nil)
	if err := fsm.Apply(context.Background(), "target_reached"); err == nil {
		t.Fatal("expected illegal transition")
	}
}

func TestFanFSM(t *testing.T) {
	f := NewFanFSM("fan-1", nil)
	ctx := context.Background()
	if err := f.Apply(ctx, "start"); err != nil {
		t.Fatal(err)
	}
	if err := f.Apply(ctx, "ramped"); err != nil {
		t.Fatal(err)
	}
	if !f.Running() {
		t.Fatal("expected running")
	}
}

func TestTowerSideEffect(t *testing.T) {
	called := false
	fsm := NewTowerFSM("tower-a1", func(ctx context.Context, tower model.TowerID, from, to model.TowerState) error {
		called = true
		return nil
	})
	if err := fsm.Apply(context.Background(), "preheat"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("side effect not called")
	}
}
