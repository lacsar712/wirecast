package tundish

import (
	"context"
	"math"

	"github.com/lacsar712/wirecast/internal/model"
)

// Controller coordinates inter-segment dehumidification vents based on
// moisture gradient and per-zone dew-point targets.
type Controller struct {
	bank       *VentBank
	maxDelta   float64
	baseOpen   float64
	maxOpen    float64
	gradientFn func([]model.MoistureReading) float64
}

func NewController(bank *VentBank, maxGradientDelta float64) *Controller {
	return &Controller{
		bank: bank, maxDelta: maxGradientDelta,
		baseOpen: 25, maxOpen: 85,
	}
}

func (c *Controller) Bank() *VentBank { return c.bank }

func (c *Controller) WithGradientFn(fn func([]model.MoistureReading) float64) {
	c.gradientFn = fn
}

func (c *Controller) PlanCycle(readings []model.MoistureReading) []VentAction {
	if c.bank == nil || c.bank.Count() == 0 {
		return nil
	}
	byZone := indexReadings(readings)
	var actions []VentAction
	for _, vent := range c.bank.All() {
		upper, uok := byZone[vent.Upper]
		lower, lok := byZone[vent.Lower]
		if !uok || !lok {
			continue
		}
		spread := upper.Pct - lower.Pct
		if spread <= 0.2 {
			if vent.Active() {
				actions = append(actions, VentAction{Vent: vent, OpenPct: 0, Reason: "gradient_flat"})
			}
			continue
		}
		open := c.baseOpen + (spread / c.maxDelta) * (c.maxOpen - c.baseOpen)
		if open > c.maxOpen {
			open = c.maxOpen
		}
		if lower.Pct > vent.DewPointTarget()+1.0 {
			open = math.Min(open+c.baseOpen, c.maxOpen)
		}
		actions = append(actions, VentAction{
			Vent: vent, OpenPct: open, Reason: "moisture_spread",
		})
	}
	return actions
}

func (c *Controller) ApplyCycle(ctx context.Context, readings []model.MoistureReading) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, act := range c.PlanCycle(readings) {
		act.Vent.Drive(act.OpenPct)
	}
	return nil
}

func (c *Controller) ValidateGradient(readings []model.MoistureReading) error {
	delta := c.gradientDelta(readings)
	if delta > c.maxDelta {
		return model.Wrap("dehumid", "gradient", model.ErrGradient)
	}
	return nil
}

func (c *Controller) gradientDelta(readings []model.MoistureReading) float64 {
	if c.gradientFn != nil {
		return c.gradientFn(readings)
	}
	if len(readings) < 2 {
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

func (c *Controller) EqualizeVents(ctx context.Context) {
	if c.bank == nil {
		return
	}
	for _, vent := range c.bank.All() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		vent.Drive(c.baseOpen)
	}
}

func (c *Controller) SealAll() {
	if c.bank != nil {
		c.bank.CloseAll()
	}
}

func indexReadings(readings []model.MoistureReading) map[model.ZoneID]model.MoistureReading {
	out := make(map[model.ZoneID]model.MoistureReading, len(readings))
	for _, r := range readings {
		out[r.Zone] = r
	}
	return out
}
