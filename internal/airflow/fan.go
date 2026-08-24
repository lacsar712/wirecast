package airflow

import (
	"context"

	"github.com/lacsar712/wirecast/internal/fsm"
	"github.com/lacsar712/wirecast/internal/model"
)

type Fan struct {
	id    model.FanID
	fsm   *fsm.FanFSM
	speed float64
}

func NewFan(id model.FanID) *Fan {
	return &Fan{id: id, fsm: fsm.NewFanFSM(id, nil)}
}

func (f *Fan) ID() model.FanID { return f.id }

func (f *Fan) Start(ctx context.Context) error {
	if err := f.fsm.Apply(ctx, "start"); err != nil {
		return err
	}
	return f.fsm.Apply(ctx, "ramped")
}

func (f *Fan) Stop(ctx context.Context) error {
	if f.fsm.State() == model.FanOff {
		return nil
	}
	if err := f.fsm.Apply(ctx, "stop"); err != nil {
		return err
	}
	return f.fsm.Apply(ctx, "coast_done")
}

func (f *Fan) SetSpeed(pct float64) {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	f.speed = pct
}

func (f *Fan) Speed() float64 { return f.speed }

func (f *Fan) State() model.FanState { return f.fsm.State() }

func (f *Fan) Running() bool { return f.fsm.Running() }

type FanBank struct {
	fans map[model.FanID]*Fan
}

func NewFanBank() *FanBank {
	return &FanBank{fans: make(map[model.FanID]*Fan)}
}

func (b *FanBank) Add(f *Fan) {
	b.fans[f.id] = f
}

func (b *FanBank) Get(id model.FanID) (*Fan, bool) {
	f, ok := b.fans[id]
	return f, ok
}

func (b *FanBank) StartAll(ctx context.Context) error {
	for _, f := range b.fans {
		if err := f.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (b *FanBank) TotalSpeed() float64 {
	var sum float64
	for _, f := range b.fans {
		sum += f.Speed()
	}
	return sum
}

func (b *FanBank) Count() int { return len(b.fans) }
