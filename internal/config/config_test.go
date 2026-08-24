package config

import "testing"

func TestDefaultValid(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestLoad(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TowerID == "" {
		t.Fatal("tower id required")
	}
}

func TestInvalidZoneCount(t *testing.T) {
	cfg := Default()
	cfg.ZoneCount = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
