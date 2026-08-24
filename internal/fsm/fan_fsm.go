package fsm

import (
	"context"

	"github.com/lacsar712/wirecast/internal/model"
)

type FanSideEffect func(ctx context.Context, fan model.FanID, from, to model.FanState) error

type FanFSM struct {
	id       model.FanID
	state    model.FanState
	onChange FanSideEffect
}

func NewFanFSM(id model.FanID, effect FanSideEffect) *FanFSM {
	return &FanFSM{id: id, state: model.FanOff, onChange: effect}
}

func (f *FanFSM) State() model.FanState { return f.state }

func (f *FanFSM) ID() model.FanID { return f.id }

func (f *FanFSM) Apply(ctx context.Context, event string) error {
	next, err := MustFan(f.state, event)
	if err != nil {
		return err
	}
	prev := f.state
	if f.onChange != nil {
		if err := f.onChange(ctx, f.id, prev, next); err != nil {
			return model.Wrap("fan_fsm", "side_effect", err)
		}
	}
	f.state = next
	return nil
}

func (f *FanFSM) Running() bool {
	return f.state == model.FanRun || f.state == model.FanRamp
}
