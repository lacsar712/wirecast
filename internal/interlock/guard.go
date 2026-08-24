package interlock

import (
	"github.com/lacsar712/wirecast/internal/model"
)

type Guard struct {
	allowed map[model.ZoneID]model.PlenumID
}

func NewGuard(pairs map[model.ZoneID]model.PlenumID) *Guard {
	cp := make(map[model.ZoneID]model.PlenumID, len(pairs))
	for k, v := range pairs {
		cp[k] = v
	}
	return &Guard{allowed: cp}
}

func (g *Guard) Permit(zone model.ZoneID, plenum model.PlenumID) error {
	want, ok := g.allowed[zone]
	if !ok {
		return model.Wrap("interlock", "unknown_zone", model.ErrNotFound)
	}
	if want != plenum {
		return model.Wrap("interlock", "plenum_mismatch", model.ErrInterlock)
	}
	return nil
}

func (g *Guard) ZonesFor(plenum model.PlenumID) []model.ZoneID {
	var out []model.ZoneID
	for zone, p := range g.allowed {
		if p == plenum {
			out = append(out, zone)
		}
	}
	return out
}

func (g *Guard) Register(zone model.ZoneID, plenum model.PlenumID) {
	g.allowed[zone] = plenum
}
