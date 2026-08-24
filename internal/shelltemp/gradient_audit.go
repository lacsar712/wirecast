package shelltemp

import (
	"sort"

	"github.com/lacsar712/wirecast/internal/model"
)

// OrderedGradientValidator checks moisture decreases monotonically from the
// top segment downward and that adjacent spread stays within limits.
type OrderedGradientValidator struct {
	order      []model.ZoneID
	maxDelta   float64
	maxReverse float64
}

func NewOrderedGradientValidator(order []model.ZoneID, maxDelta, maxReverse float64) *OrderedGradientValidator {
	cp := make([]model.ZoneID, len(order))
	copy(cp, order)
	return &OrderedGradientValidator{
		order: cp, maxDelta: maxDelta, maxReverse: maxReverse,
	}
}

func (v *OrderedGradientValidator) ValidateOrdered(readings []model.MoistureReading) error {
	byZone := indexByZone(readings)
	var prev float64
	first := true
	for _, zone := range v.order {
		r, ok := byZone[zone]
		if !ok {
			continue
		}
		if !first {
			step := prev - r.Pct
			if step < -v.maxReverse {
				return model.Wrap("gradient_audit", zone.String(), model.ErrGradient)
			}
			if step > v.maxDelta {
				return model.Wrap("gradient_audit", "segment_spread", model.ErrGradient)
			}
		}
		prev = r.Pct
		first = false
	}
	return nil
}

func (v *OrderedGradientValidator) AuditReport(readings []model.MoistureReading) GradientAuditReport {
	byZone := indexByZone(readings)
	report := GradientAuditReport{Zones: len(v.order)}
	var prev float64
	first := true
	for i, zone := range v.order {
		r, ok := byZone[zone]
		if !ok {
			continue
		}
		sample := SegmentSample{Zone: zone, Reading: r.Pct, Index: i}
		if !first {
			sample.StepDelta = prev - r.Pct
			if sample.StepDelta < -v.maxReverse {
				report.ReverseViolations++
			}
			if sample.StepDelta > v.maxDelta {
				report.SpreadViolations++
			}
		}
		report.Samples = append(report.Samples, sample)
		prev = r.Pct
		first = false
	}
	sort.Slice(report.Samples, func(i, j int) bool {
		return report.Samples[i].Index < report.Samples[j].Index
	})
	return report
}

func (v *OrderedGradientValidator) WorstSpread(readings []model.MoistureReading) float64 {
	byZone := indexByZone(readings)
	var worst float64
	var prev float64
	first := true
	for _, zone := range v.order {
		r, ok := byZone[zone]
		if !ok {
			continue
		}
		if !first {
			step := prev - r.Pct
			if step < 0 {
				step = -step
			}
			if step > worst {
				worst = step
			}
		}
		prev = r.Pct
		first = false
	}
	return worst
}

func (v *OrderedGradientValidator) Passes(readings []model.MoistureReading) bool {
	return v.ValidateOrdered(readings) == nil
}

type SegmentSample struct {
	Zone      model.ZoneID
	Reading   float64
	Index     int
	StepDelta float64
}

type GradientAuditReport struct {
	Zones              int
	Samples            []SegmentSample
	ReverseViolations  int
	SpreadViolations   int
}

func (r GradientAuditReport) Clean() bool {
	return r.ReverseViolations == 0 && r.SpreadViolations == 0
}

func indexByZone(readings []model.MoistureReading) map[model.ZoneID]model.MoistureReading {
	out := make(map[model.ZoneID]model.MoistureReading, len(readings))
	for _, rd := range readings {
		out[rd.Zone] = rd
	}
	return out
}
