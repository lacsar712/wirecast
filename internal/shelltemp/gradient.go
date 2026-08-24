package shelltemp

import (
	"math"

	"github.com/lacsar712/wirecast/internal/model"
)

type GradientController struct {
	maxDeltaPct float64
}

func NewGradientController(maxDeltaPct float64) *GradientController {
	return &GradientController{maxDeltaPct: maxDeltaPct}
}

func (g *GradientController) MaxDelta() float64 { return g.maxDeltaPct }

func (g *GradientController) Validate(readings []model.MoistureReading) error {
	if len(readings) < 2 {
		return nil
	}
	min, max := readings[0].Pct, readings[0].Pct
	for _, r := range readings[1:] {
		if r.Pct < min {
			min = r.Pct
		}
		if r.Pct > max {
			max = r.Pct
		}
	}
	delta := max - min
	if delta > g.maxDeltaPct {
		return model.Wrap("gradient", "exceeds_max", model.ErrGradient)
	}
	return nil
}

func (g *GradientController) Delta(readings []model.MoistureReading) float64 {
	if len(readings) == 0 {
		return 0
	}
	min, max := readings[0].Pct, readings[0].Pct
	for _, r := range readings[1:] {
		if r.Pct < min {
			min = r.Pct
		}
		if r.Pct > max {
			max = r.Pct
		}
	}
	return max - min
}

func (g *GradientController) SuggestAdjustments(readings []model.MoistureReading, profile model.MoistureProfile) map[model.ZoneID]float64 {
	out := make(map[model.ZoneID]float64)
	if len(readings) == 0 || len(profile.Targets) == 0 {
		return out
	}
	zoneTarget := make(map[model.ZoneID]float64)
	for i, z := range profile.Zones {
		if i < len(profile.Targets) {
			zoneTarget[z] = profile.Targets[i]
		}
	}
	for _, r := range readings {
		target, ok := zoneTarget[r.Zone]
		if !ok {
			continue
		}
		delta := r.Pct - target
		if math.Abs(delta) > 0.1 {
			out[r.Zone] = -delta * 0.5
		}
	}
	return out
}

func (g *GradientController) Samples(readings []model.MoistureReading) []model.GradientSample {
	out := make([]model.GradientSample, len(readings))
	for i, r := range readings {
		out[i] = model.GradientSample{Zone: r.Zone, Reading: r.Pct, At: r.At}
	}
	return out
}
