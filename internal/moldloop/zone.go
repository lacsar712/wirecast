package moldloop

import (
	"context"

	"github.com/lacsar712/wirecast/internal/model"
)

type ZoneController struct {
	zone   model.ZoneID
	plenum model.PlenumID
	enabled bool
}

func NewZoneController(zone model.ZoneID, plenum model.PlenumID) *ZoneController {
	return &ZoneController{zone: zone, plenum: plenum, enabled: true}
}

func (z *ZoneController) Zone() model.ZoneID { return z.zone }

func (z *ZoneController) Plenum() model.PlenumID { return z.plenum }

func (z *ZoneController) Enabled() bool { return z.enabled }

func (z *ZoneController) Enable() { z.enabled = true }

func (z *ZoneController) Disable() { z.enabled = false }

func (z *ZoneController) PermitAirflow(guard interface {
	Permit(zone model.ZoneID, plenum model.PlenumID) error
}) error {
	if !z.enabled {
		return model.Wrap("zone", "disabled", model.ErrConflict)
	}
	return guard.Permit(z.zone, z.plenum)
}

func (z *ZoneController) ApplyMoistureAdjustment(ctx context.Context, adj float64, plant *CasterPlant) error {
	if adj == 0 {
		return nil
	}
	var curPct float64
	for _, r := range plant.sensors.Readings() {
		if r.Zone == z.zone {
			curPct = r.Pct
			break
		}
	}
	newPct := curPct + adj
	if newPct < 0 {
		newPct = 0
	}
	return plant.ObserveMoisture(z.zone, newPct)
}

type ZoneRegistry struct {
	zones map[model.ZoneID]*ZoneController
}

func NewZoneRegistry() *ZoneRegistry {
	return &ZoneRegistry{zones: make(map[model.ZoneID]*ZoneController)}
}

func (r *ZoneRegistry) Register(z *ZoneController) {
	r.zones[z.zone] = z
}

func (r *ZoneRegistry) Get(zone model.ZoneID) (*ZoneController, bool) {
	z, ok := r.zones[zone]
	return z, ok
}

func (r *ZoneRegistry) All() []*ZoneController {
	out := make([]*ZoneController, 0, len(r.zones))
	for _, z := range r.zones {
		out = append(out, z)
	}
	return out
}

func (r *ZoneRegistry) EnabledCount() int {
	n := 0
	for _, z := range r.zones {
		if z.enabled {
			n++
		}
	}
	return n
}
