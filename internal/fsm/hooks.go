package fsm

import (
	"context"

	"github.com/lacsar712/wirecast/internal/model"
)

type DryTransitionHook func(ctx context.Context, from, to model.DryState, event string) error

type DryHookChain struct {
	before []DryTransitionHook
	after  []DryTransitionHook
}

func NewDryHookChain() *DryHookChain { return &DryHookChain{} }

func (h *DryHookChain) OnBefore(fn DryTransitionHook) { h.before = append(h.before, fn) }
func (h *DryHookChain) OnAfter(fn DryTransitionHook)  { h.after = append(h.after, fn) }

func (h *DryHookChain) RunBefore(ctx context.Context, from, to model.DryState, event string) error {
	for _, fn := range h.before {
		if err := fn(ctx, from, to, event); err != nil {
			return err
		}
	}
	return nil
}

func (h *DryHookChain) RunAfter(ctx context.Context, from, to model.DryState, event string) error {
	for _, fn := range h.after {
		if err := fn(ctx, from, to, event); err != nil {
			return err
		}
	}
	return nil
}

// DryHeatPulse counts heater drive side effects from CasterFSM after hooks (acceptance).
var DryHeatPulse func()

func RegisterDryHeatHook(chain *DryHookChain) {
	chain.OnAfter(func(ctx context.Context, from, to model.DryState, event string) error {
		_ = ctx
		_ = from
		_ = to
		_ = event
		if DryHeatPulse != nil {
			DryHeatPulse()
		}
		return nil
	})
}
