package tundish

import (
	"fmt"

	"github.com/lacsar712/wirecast/internal/model"
)

// SegmentVent models an inter-segment dehumidification vent between two
// adjacent drying zones. Vents bleed humid exhaust from the lower segment
// while the upper segment remains under positive plenum pressure.
type SegmentVent struct {
	Upper    model.ZoneID
	Lower    model.ZoneID
	Damper   model.DamperID
	openPct  float64
	active   bool
	dewPoint float64
}

func NewSegmentVent(upper, lower model.ZoneID, damper model.DamperID) *SegmentVent {
	return &SegmentVent{
		Upper: upper, Lower: lower, Damper: damper,
		dewPoint: 12.0,
	}
}

func (v *SegmentVent) OpenPct() float64 { return v.openPct }

func (v *SegmentVent) Active() bool { return v.active }

func (v *SegmentVent) DewPointTarget() float64 { return v.dewPoint }

func (v *SegmentVent) SetDewPointTarget(pct float64) {
	if pct < 0 {
		pct = 0
	}
	v.dewPoint = pct
}

func (v *SegmentVent) Drive(openPct float64) {
	if openPct <= 0 {
		v.openPct = 0
		v.active = false
		return
	}
	if openPct > 100 {
		openPct = 100
	}
	v.openPct = openPct
	v.active = true
}

func (v *SegmentVent) Close() {
	v.openPct = 0
	v.active = false
}

// VentBank tracks all inter-segment vents for a tower ordered top-to-bottom.
type VentBank struct {
	vents []*SegmentVent
}

func NewVentBank(zones []model.ZoneID, damperPrefix string) (*VentBank, error) {
	if len(zones) < 2 {
		return &VentBank{}, nil
	}
	b := &VentBank{vents: make([]*SegmentVent, 0, len(zones)-1)}
	for i := 0; i < len(zones)-1; i++ {
		damperID, err := model.ParseDamperID(fmt.Sprintf("%s-%02d", damperPrefix, i))
		if err != nil {
			return nil, err
		}
		b.vents = append(b.vents, NewSegmentVent(zones[i], zones[i+1], damperID))
	}
	return b, nil
}

func (b *VentBank) Count() int { return len(b.vents) }

func (b *VentBank) All() []*SegmentVent {
	out := make([]*SegmentVent, len(b.vents))
	copy(out, b.vents)
	return out
}

func (b *VentBank) Between(upper, lower model.ZoneID) (*SegmentVent, bool) {
	for _, v := range b.vents {
		if v.Upper == upper && v.Lower == lower {
			return v, true
		}
	}
	return nil, false
}

func (b *VentBank) ActiveCount() int {
	n := 0
	for _, v := range b.vents {
		if v.Active() {
			n++
		}
	}
	return n
}

func (b *VentBank) TotalOpenPct() float64 {
	var sum float64
	for _, v := range b.vents {
		sum += v.OpenPct()
	}
	return sum
}

func (b *VentBank) CloseAll() {
	for _, v := range b.vents {
		v.Close()
	}
}

// VentAction describes a commanded vent position change.
type VentAction struct {
	Vent    *SegmentVent
	OpenPct float64
	Reason  string
}
