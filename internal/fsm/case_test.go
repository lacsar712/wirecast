package fsm

import (
	"context"
	"testing"

	"github.com/lacsar712/wirecast/internal/model"
)

func TestCase(t *testing.T) {
	DryHeatPulse = nil
	var pulses int
	DryHeatPulse = func() { pulses++ }
	d := NewDryFSM(model.TowerID("tower-a1"), nil)
	RegisterDryHeatHook(d.Hooks())
	_, err := d.Dispatch(context.Background(), "hold")
	if err == nil {
		t.Fatal("expected illegal dry transition error")
	}
	if pulses != 0 {
		t.Fatalf("illegal dry transition should not pulse heat drive, got %d", pulses)
	}
}
