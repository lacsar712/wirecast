package airflow

import (
	"context"
	"time"

	"github.com/lacsar712/wirecast/internal/clock"
	"github.com/lacsar712/wirecast/internal/model"
)

type Plenum struct {
	ID       model.PlenumID
	Capacity float64
	setpoint model.AirflowSetpoint
	actual   float64
	primed   bool
}

func NewPlenum(id model.PlenumID, capacity float64) *Plenum {
	return &Plenum{ID: id, Capacity: capacity}
}

func (p *Plenum) BindSetpoint(sp model.AirflowSetpoint) {
	p.setpoint = sp
}

func (p *Plenum) Setpoint() model.AirflowSetpoint { return p.setpoint }

func (p *Plenum) ObserveFlow(cmh float64) {
	p.actual = cmh
}

func (p *Plenum) ActualFlow() float64 { return p.actual }

func (p *Plenum) Primed() bool { return p.primed }

func (p *Plenum) Prime(ctx context.Context, clk clock.Clock, duration time.Duration) error {
	deadline := clk.Now().Add(duration)
	for clk.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}
		p.primed = true
		if pc, ok := clk.(*clock.ProcessClock); ok {
			pc.Step()
		} else {
			time.Sleep(5 * time.Millisecond)
		}
	}
	return nil
}

func (p *Plenum) WithinSetpoint() bool {
	return p.setpoint.Within(p.actual)
}

type PlenumTable struct {
	plenums map[model.PlenumID]*Plenum
}

func NewPlenumTable() *PlenumTable {
	return &PlenumTable{plenums: make(map[model.PlenumID]*Plenum)}
}

func (t *PlenumTable) Add(p *Plenum) {
	t.plenums[p.ID] = p
}

func (t *PlenumTable) Get(id model.PlenumID) (*Plenum, bool) {
	p, ok := t.plenums[id]
	return p, ok
}

func (t *PlenumTable) ValidateAll() error {
	for id, p := range t.plenums {
		if p.setpoint.CubicMetersPerHour > 0 && !p.WithinSetpoint() {
			return model.Wrap("plenum", id.String(), model.ErrAirflowSetpoint)
		}
	}
	return nil
}

func (t *PlenumTable) List() []*Plenum {
	out := make([]*Plenum, 0, len(t.plenums))
	for _, p := range t.plenums {
		out = append(out, p)
	}
	return out
}

type Router struct {
	routes []model.PlenumRoute
}

func NewRouter(routes []model.PlenumRoute) *Router {
	cp := make([]model.PlenumRoute, len(routes))
	copy(cp, routes)
	return &Router{routes: cp}
}

func (r *Router) Route(from model.PlenumID) (model.PlenumRoute, bool) {
	var best model.PlenumRoute
	found := false
	for _, route := range r.routes {
		if route.From != from {
			continue
		}
		if !found || route.Priority > best.Priority {
			best = route
			found = true
		}
	}
	return best, found
}

func (r *Router) RoutesFrom(from model.PlenumID) []model.PlenumRoute {
	var out []model.PlenumRoute
	for _, route := range r.routes {
		if route.From == from {
			out = append(out, route)
		}
	}
	return out
}
