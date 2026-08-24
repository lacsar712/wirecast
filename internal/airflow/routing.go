package airflow

import (
	"context"
	"sort"

	"github.com/lacsar712/wirecast/internal/model"
)

// RoutePlanner selects plenum paths and distributes airflow across zones.
type RoutePlanner struct {
	router  *Router
	plenums *PlenumTable
}

func NewRoutePlanner(router *Router, plenums *PlenumTable) *RoutePlanner {
	return &RoutePlanner{router: router, plenums: plenums}
}

func (p *RoutePlanner) SelectPath(from model.PlenumID, demand float64) (model.PlenumRoute, error) {
	routes := p.router.RoutesFrom(from)
	if len(routes) == 0 {
		return model.PlenumRoute{}, model.Wrap("route_planner", "no_route", model.ErrNotFound)
	}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Priority != routes[j].Priority {
			return routes[i].Priority > routes[j].Priority
		}
		return routes[i].Damper < routes[j].Damper
	})
	best := routes[0]
	if pl, ok := p.plenums.Get(best.To); ok && pl.Capacity > 0 && demand > pl.Capacity {
		for _, route := range routes[1:] {
			if alt, ok := p.plenums.Get(route.To); ok && alt.Capacity >= demand {
				return route, nil
			}
		}
	}
	return best, nil
}

// FlowAllocation maps each zone to its assigned cubic-meters-per-hour share.
type FlowAllocation map[model.ZoneID]float64

func (p *RoutePlanner) AllocateFlow(plenum model.PlenumID, zones []model.ZoneID, totalCMH float64) FlowAllocation {
	out := make(FlowAllocation, len(zones))
	if len(zones) == 0 || totalCMH <= 0 {
		return out
	}
	share := totalCMH / float64(len(zones))
	for _, z := range zones {
		out[z] = share
	}
	pl, ok := p.plenums.Get(plenum)
	if !ok || pl.Capacity <= 0 {
		return out
	}
	scale := 1.0
	if totalCMH > pl.Capacity {
		scale = pl.Capacity / totalCMH
	}
	for z, v := range out {
		out[z] = v * scale
	}
	return out
}

func (p *RoutePlanner) ValidateRouting(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, pl := range p.plenums.List() {
		if pl.setpoint.CubicMetersPerHour > 0 && pl.Capacity > 0 {
			if pl.setpoint.CubicMetersPerHour > pl.Capacity {
				return model.Wrap("route_planner", pl.ID.String(), model.ErrAirflowSetpoint)
			}
		}
	}
	return nil
}

func (p *RoutePlanner) BindRoute(ctx context.Context, from, to model.PlenumID, damper model.DamperID, sp model.AirflowSetpoint) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	route, err := p.SelectPath(from, sp.CubicMetersPerHour)
	if err != nil {
		return err
	}
	if route.To != to && to != "" {
		return model.Wrap("route_planner", "path_mismatch", model.ErrInterlock)
	}
	if pl, ok := p.plenums.Get(from); ok {
		pl.BindSetpoint(sp)
	}
	_ = damper
	return nil
}

func (p *RoutePlanner) Summarize(from model.PlenumID) RouteSummary {
	routes := p.router.RoutesFrom(from)
	summary := RouteSummary{From: from, RouteCount: len(routes)}
	for _, r := range routes {
		entry := RouteEntry{Route: r}
		if pl, ok := p.plenums.Get(r.To); ok {
			entry.Capacity = pl.Capacity
			entry.Actual = pl.actual
		}
		summary.Entries = append(summary.Entries, entry)
	}
	return summary
}

type RouteEntry struct {
	Route    model.PlenumRoute
	Capacity float64
	Actual   float64
}

type RouteSummary struct {
	From       model.PlenumID
	RouteCount int
	Entries    []RouteEntry
}

func (p *RoutePlanner) ApplyAllocation(alloc FlowAllocation, zones *ZoneFlowTable) {
	for zone, cmh := range alloc {
		zones.Set(zone, cmh)
	}
}

// ZoneFlowTable tracks last-known per-zone airflow from plenum routing.
type ZoneFlowTable struct {
	flows map[model.ZoneID]float64
}

func NewZoneFlowTable() *ZoneFlowTable {
	return &ZoneFlowTable{flows: make(map[model.ZoneID]float64)}
}

func (t *ZoneFlowTable) Set(zone model.ZoneID, cmh float64) {
	t.flows[zone] = cmh
}

func (t *ZoneFlowTable) Get(zone model.ZoneID) (float64, bool) {
	v, ok := t.flows[zone]
	return v, ok
}

func (t *ZoneFlowTable) Total() float64 {
	var sum float64
	for _, v := range t.flows {
		sum += v
	}
	return sum
}
