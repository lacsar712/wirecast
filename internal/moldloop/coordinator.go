package moldloop

import (
	"context"
	"time"

	"github.com/lacsar712/wirecast/internal/airflow"
	"github.com/lacsar712/wirecast/internal/clock"
	"github.com/lacsar712/wirecast/internal/config"
	"github.com/lacsar712/wirecast/internal/model"
)

type Coordinator struct {
	cfg  config.Config
	clk  clock.Clock
	fans *airflow.FanBank
}

func NewCoordinator(cfg config.Config, clk clock.Clock, fans *airflow.FanBank) *Coordinator {
	return &Coordinator{cfg: cfg, clk: clk, fans: fans}
}

func (c *Coordinator) Start(ctx context.Context, fan model.FanID) error {
	f, ok := c.fans.Get(fan)
	if !ok {
		return model.Wrap("coordinator", "fan", model.ErrNotFound)
	}
	if err := f.Start(ctx); err != nil {
		return err
	}
	f.SetSpeed(100)
	return nil
}

func (c *Coordinator) Stop(ctx context.Context, fan model.FanID) error {
	f, ok := c.fans.Get(fan)
	if !ok {
		return model.Wrap("coordinator", "fan", model.ErrNotFound)
	}
	return f.Stop(ctx)
}

func (c *Coordinator) RampTo(ctx context.Context, fan model.FanID, targetPct float64, steps int) error {
	f, ok := c.fans.Get(fan)
	if !ok {
		return model.Wrap("coordinator", "ramp", model.ErrNotFound)
	}
	if steps <= 0 {
		steps = 1
	}
	cur := f.Speed()
	delta := (targetPct - cur) / float64(steps)
	for i := 0; i < steps; i++ {
		cur += delta
		f.SetSpeed(cur)
		if pc, ok := c.clk.(*clock.ProcessClock); ok {
			pc.Step()
		}
	}
	return nil
}

func (c *Coordinator) MinRunElapsed(fan model.FanID) bool {
	f, ok := c.fans.Get(fan)
	if !ok {
		return false
	}
	return f.Running() && c.cfg.FanMinRun > 0
}

func (c *Coordinator) CoastWindow() time.Duration {
	return c.cfg.FanCoast
}

func (c *Coordinator) WaitCoast(ctx context.Context) error {
	if c.cfg.FanCoast <= 0 {
		return nil
	}
	if pc, ok := c.clk.(*clock.ProcessClock); ok {
		pc.Advance(c.cfg.FanCoast)
		return nil
	}
	deadline := c.clk.Now().Add(c.cfg.FanCoast)
	return clock.WaitUntilContext(ctx, c.clk, deadline)
}
