package store

import (
	"testing"
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

func TestMemoryTower(t *testing.T) {
	mem := NewMemory()
	snap := NewSnapshotBuilder("tower-a1").State(model.TowerDrying).Build(time.Now())
	mem.PutTower(snap)
	got, ok := mem.GetTower("tower-a1")
	if !ok || got.State != model.TowerDrying {
		t.Fatal("tower not stored")
	}
}

func TestScheduleStore(t *testing.T) {
	mem := NewMemory()
	ss := NewScheduleStore(mem)
	now := time.Date(2024, 8, 1, 12, 0, 0, 0, time.UTC)
	sched := model.DryingSchedule{
		ID: "sch1",
		Entries: []model.DryingScheduleEntry{{
			ID: "e1", Start: now.Add(-time.Hour), End: now.Add(time.Hour),
			Plenum: "plenum-main", TargetMoistPct: 14,
		}},
	}
	ss.Save(sched)
	entry, ok := ss.ActiveEntry(sched, now)
	if !ok || entry.TargetMoistPct != 14 {
		t.Fatal("active entry mismatch")
	}
}

func TestDiffZones(t *testing.T) {
	before := model.TowerSnapshot{
		Zones: []model.ZoneAssignment{{Zone: "z1", ActualMoist: 18}},
	}
	after := model.TowerSnapshot{
		Zones: []model.ZoneAssignment{{Zone: "z1", ActualMoist: 15}},
	}
	changed := DiffZones(before, after)
	if len(changed) != 1 {
		t.Fatal("expected change")
	}
}

func TestCloneTowerSnapshot(t *testing.T) {
	orig := model.TowerSnapshot{
		ID: "t1", Zones: []model.ZoneAssignment{{Zone: "z1", ActualMoist: 20}},
	}
	cl := CloneTowerSnapshot(orig)
	orig.Zones[0].ActualMoist = 99
	if cl.Zones[0].ActualMoist == 99 {
		t.Fatal("clone should be independent")
	}
}
