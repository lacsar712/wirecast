package interlock

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

func TestGuardPermit(t *testing.T) {
	g := NewGuard(map[model.ZoneID]model.PlenumID{
		"zone-1": "plenum-main",
	})
	if err := g.Permit("zone-1", "plenum-main"); err != nil {
		t.Fatal(err)
	}
	if err := g.Permit("zone-1", "plenum-alt"); err == nil {
		t.Fatal("expected mismatch")
	}
}

func TestGuardZonesFor(t *testing.T) {
	g := NewGuard(map[model.ZoneID]model.PlenumID{
		"zone-1": "plenum-main",
		"zone-2": "plenum-main",
		"zone-3": "plenum-alt",
	})
	zones := g.ZonesFor("plenum-main")
	if len(zones) != 2 {
		t.Fatalf("got %d zones", len(zones))
	}
}

func TestDamperLock(t *testing.T) {
	now := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	lock := NewDamperLock(func() time.Time { return now })
	release, ok := lock.TryAcquire("damper-1", time.Minute)
	if !ok {
		t.Fatal("expected acquire")
	}
	if !lock.Held("damper-1") {
		t.Fatal("expected held")
	}
	release()
	if lock.Held("damper-1") {
		t.Fatal("expected released")
	}
}

func TestDamperLockWithLease(t *testing.T) {
	now := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	lock := NewDamperLock(func() time.Time { return now })
	err := lock.WithLease(context.Background(), "damper-2", time.Minute, func() error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
