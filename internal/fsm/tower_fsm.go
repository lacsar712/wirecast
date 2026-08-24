package fsm

import (
	"context"

	"github.com/lacsar712/wirecast/internal/model"
)

type TowerSideEffect func(ctx context.Context, tower model.TowerID, from, to model.TowerState) error

type TowerFSM struct {
	id       model.TowerID
	state    model.TowerState
	onChange TowerSideEffect
}

func NewTowerFSM(id model.TowerID, effect TowerSideEffect) *TowerFSM {
	return &TowerFSM{id: id, state: model.TowerIdle, onChange: effect}
}

func (f *TowerFSM) State() model.TowerState { return f.state }

func (f *TowerFSM) ID() model.TowerID { return f.id }

func (f *TowerFSM) Apply(ctx context.Context, event string) error {
	next, err := MustTower(f.state, event)
	if err != nil {
		return err
	}
	prev := f.state
	if f.onChange != nil {
		if err := f.onChange(ctx, f.id, prev, next); err != nil {
			return model.Wrap("tower_fsm", "side_effect", err)
		}
	}
	f.state = next
	return nil
}

func (f *TowerFSM) ForceState(s model.TowerState) {
	f.state = s
}
