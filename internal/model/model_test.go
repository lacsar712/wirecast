package model

import "testing"

func TestAirflowWithin(t *testing.T) {
	sp := AirflowSetpoint{CubicMetersPerHour: 100, TolerancePct: 10}
	if !sp.Within(105) {
		t.Fatal("expected within")
	}
	if sp.Within(120) {
		t.Fatal("expected outside")
	}
}

func TestDryingScheduleClone(t *testing.T) {
	s := DryingSchedule{
		ID: "sch1",
		Entries: []DryingScheduleEntry{{ID: "e1", TargetMoistPct: 14}},
		Version: 2,
	}
	c := s.Clone()
	if len(c.Entries) != 1 || c.Entries[0].TargetMoistPct != 14 {
		t.Fatal("clone mismatch")
	}
	s.Entries[0].TargetMoistPct = 99
	if c.Entries[0].TargetMoistPct == 99 {
		t.Fatal("clone should be independent")
	}
}

func TestParseTowerID(t *testing.T) {
	id, err := ParseTowerID("tower-a1")
	if err != nil || id != "tower-a1" {
		t.Fatal(err)
	}
	if _, err := ParseTowerID(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseZoneID(t *testing.T) {
	z, err := ParseZoneID("tower-a1", 3)
	if err != nil || z.String() != "tower-a1-zone-03" {
		t.Fatal(err, z)
	}
}
