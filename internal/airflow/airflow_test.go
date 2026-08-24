package airflow

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/wirecast/internal/clock"
	"github.com/lacsar712/wirecast/internal/interlock"
	"github.com/lacsar712/wirecast/internal/model"
)

func TestFanLifecycle(t *testing.T) {
	f := NewFan("fan-1")
	ctx := context.Background()
	if err := f.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if !f.Running() {
		t.Fatal("expected running")
	}
	f.SetSpeed(75)
	if f.Speed() != 75 {
		t.Fatal("speed mismatch")
	}
	if err := f.Stop(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestFanBank(t *testing.T) {
	bank := NewFanBank()
	bank.Add(NewFan("fan-1"))
	bank.Add(NewFan("fan-2"))
	if err := bank.StartAll(context.Background()); err != nil {
		t.Fatal(err)
	}
	if bank.Count() != 2 {
		t.Fatal("count")
	}
}

func TestDamperActuator(t *testing.T) {
	lock := interlock.NewDamperLock(func() time.Time {
		return time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	})
	act := NewDamperActuator(lock)
	d := NewDamper("damper-1")
	act.Register(d)
	if err := act.Move(context.Background(), "damper-1", 50); err != nil {
		t.Fatal(err)
	}
	got, ok := act.Get("damper-1")
	if !ok || got.Position() != model.DamperThrottled {
		t.Fatal("damper position")
	}
}

func TestPlenumPrime(t *testing.T) {
	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewProcessClock(start, time.Millisecond)
	p := NewPlenum("plenum-main", 1000)
	p.BindSetpoint(model.AirflowSetpoint{CubicMetersPerHour: 850, TolerancePct: 5})
	if err := p.Prime(context.Background(), clk, time.Second); err != nil {
		t.Fatal(err)
	}
	if !p.Primed() {
		t.Fatal("expected primed")
	}
	p.ObserveFlow(850)
	if !p.WithinSetpoint() {
		t.Fatal("within setpoint")
	}
}

func TestRouter(t *testing.T) {
	r := NewRouter([]model.PlenumRoute{{
		From: "plenum-main", To: "plenum-alt", Damper: "damper-1", Priority: 10,
	}})
	route, ok := r.Route("plenum-main")
	if !ok || route.Damper != "damper-1" {
		t.Fatal("route mismatch")
	}
}
