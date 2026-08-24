package airflow

import (
	"context"

	"github.com/lacsar712/wirecast/internal/interlock"
	"github.com/lacsar712/wirecast/internal/model"
)

type Damper struct {
	id       model.DamperID
	position model.DamperPosition
	openPct  float64
}

func NewDamper(id model.DamperID) *Damper {
	return &Damper{id: id, position: model.DamperClosed, openPct: 0}
}

func (d *Damper) ID() model.DamperID { return d.id }

func (d *Damper) Position() model.DamperPosition { return d.position }

func (d *Damper) OpenPct() float64 { return d.openPct }

func (d *Damper) SetOpen(pct float64) {
	if pct <= 0 {
		d.position = model.DamperClosed
		d.openPct = 0
		return
	}
	if pct >= 100 {
		d.position = model.DamperOpen
		d.openPct = 100
		return
	}
	d.position = model.DamperThrottled
	d.openPct = pct
}

type DamperActuator struct {
	lock   *interlock.DamperLock
	dampers map[model.DamperID]*Damper
}

func NewDamperActuator(lock *interlock.DamperLock) *DamperActuator {
	return &DamperActuator{lock: lock, dampers: make(map[model.DamperID]*Damper)}
}

func (a *DamperActuator) Register(d *Damper) {
	a.dampers[d.id] = d
}

func (a *DamperActuator) Move(ctx context.Context, id model.DamperID, pct float64) error {
	return a.lock.WithLease(ctx, id, interlock.DefaultLeaseTTL, func() error {
		d, ok := a.dampers[id]
		if !ok {
			return model.Wrap("damper", "unknown", model.ErrNotFound)
		}
		d.SetOpen(pct)
		return nil
	})
}

func (a *DamperActuator) Get(id model.DamperID) (*Damper, bool) {
	d, ok := a.dampers[id]
	return d, ok
}
