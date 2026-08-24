package store

import (
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

type SnapshotBuilder struct {
	id    model.TowerID
	state model.TowerState
	zones []model.ZoneAssignment
	fans  []model.FanID
}

func NewSnapshotBuilder(id model.TowerID) *SnapshotBuilder {
	return &SnapshotBuilder{id: id, state: model.TowerIdle}
}

func (b *SnapshotBuilder) State(s model.TowerState) *SnapshotBuilder {
	b.state = s
	return b
}

func (b *SnapshotBuilder) Zone(a model.ZoneAssignment) *SnapshotBuilder {
	b.zones = append(b.zones, a)
	return b
}

func (b *SnapshotBuilder) Fan(id model.FanID) *SnapshotBuilder {
	b.fans = append(b.fans, id)
	return b
}

func (b *SnapshotBuilder) Build(at time.Time) model.TowerSnapshot {
	zones := make([]model.ZoneAssignment, len(b.zones))
	copy(zones, b.zones)
	fans := make([]model.FanID, len(b.fans))
	copy(fans, b.fans)
	return model.TowerSnapshot{ID: b.id, State: b.state, Zones: zones, Fans: fans, UpdatedAt: at}
}

func DiffZones(before, after model.TowerSnapshot) []model.ZoneID {
	index := make(map[model.ZoneID]model.ZoneAssignment)
	for _, z := range before.Zones {
		index[z.Zone] = z
	}
	var changed []model.ZoneID
	for _, z := range after.Zones {
		prev, ok := index[z.Zone]
		if !ok || prev.LastFlow != z.LastFlow || prev.Enabled != z.Enabled || prev.ActualMoist != z.ActualMoist {
			changed = append(changed, z.Zone)
		}
	}
	return changed
}

func CloneTowerSnapshot(s model.TowerSnapshot) model.TowerSnapshot {
	out := model.TowerSnapshot{ID: s.ID, State: s.State, UpdatedAt: s.UpdatedAt}
	if len(s.Zones) > 0 {
		out.Zones = make([]model.ZoneAssignment, len(s.Zones))
		copy(out.Zones, s.Zones)
	}
	if len(s.Fans) > 0 {
		out.Fans = make([]model.FanID, len(s.Fans))
		copy(out.Fans, s.Fans)
	}
	return out
}

type SegmentSnapshot struct {
	Zone  model.ZoneID
	TempC float64
}

type CastSnapshot struct {
	Tower    model.TowerID
	Segments []SegmentSnapshot
}

func CloneCastSnapshot(s CastSnapshot) CastSnapshot {
	out := CastSnapshot{Tower: s.Tower}
	out.Segments = s.Segments
	return out
}
