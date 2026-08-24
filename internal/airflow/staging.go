package airflow

import (
	"context"
	"sort"

	"github.com/lacsar712/wirecast/internal/clock"
	"github.com/lacsar712/wirecast/internal/model"
)

// FanStage describes one step in a multi-fan staging sequence.
type FanStage struct {
	Speeds map[model.FanID]float64
	Hold   int // process-clock steps between stages
}

// StagePlan is an ordered fan staging ramp used during plenum priming.
type StagePlan struct {
	Stages []FanStage
}

func (p StagePlan) StageCount() int { return len(p.Stages) }

// Stager executes fan staging plans against a fan bank.
type Stager struct {
	bank *FanBank
	clk  clock.Clock
}

func NewStager(bank *FanBank, clk clock.Clock) *Stager {
	return &Stager{bank: bank, clk: clk}
}

func (s *Stager) Execute(ctx context.Context, plan StagePlan) error {
	if len(plan.Stages) == 0 {
		return model.Wrap("stager", "empty_plan", model.ErrConflict)
	}
	for _, stage := range plan.Stages {
		if err := ctx.Err(); err != nil {
			return err
		}
		for fanID, pct := range stage.Speeds {
			f, ok := s.bank.Get(fanID)
			if !ok {
				return model.Wrap("stager", "fan", model.ErrNotFound)
			}
			if f.State() == model.FanOff {
				if err := f.Start(ctx); err != nil {
					return err
				}
			}
			f.SetSpeed(pct)
		}
		hold := stage.Hold
		if hold <= 0 {
			hold = 1
		}
		if pc, ok := s.clk.(*clock.ProcessClock); ok {
			for i := 0; i < hold; i++ {
				pc.Step()
			}
		}
	}
	return nil
}

func (s *Stager) CurrentLoad() float64 {
	return s.bank.TotalSpeed()
}

func (s *Stager) RunningCount() int {
	n := 0
	for _, f := range s.bankList() {
		if f.Running() {
			n++
		}
	}
	return n
}

func (s *Stager) bankList() []*Fan {
	var out []*Fan
	for _, f := range s.bank.fans {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID() < out[j].ID()
	})
	return out
}

// BuildStagingPlan constructs a fan staging plan from zone demand and installed fans.
func BuildStagingPlan(zoneCount int, targetCMH float64, fanIDs []model.FanID) StagePlan {
	if len(fanIDs) == 0 {
		return StagePlan{}
	}
	sort.Slice(fanIDs, func(i, j int) bool { return fanIDs[i] < fanIDs[j] })
	perFan := targetCMH / float64(len(fanIDs))
	basePct := fanPctFromCMH(perFan)
	stages := make([]FanStage, 0, len(fanIDs)+2)
	ramp := []float64{0.35, 0.65, 1.0}
	for _, factor := range ramp {
		speeds := make(map[model.FanID]float64, len(fanIDs))
		for _, id := range fanIDs {
			speeds[id] = basePct * factor
		}
		stages = append(stages, FanStage{Speeds: speeds, Hold: zoneCount})
	}
	return StagePlan{Stages: stages}
}

func fanPctFromCMH(cmh float64) float64 {
	if cmh <= 0 {
		return 0
	}
	pct := (cmh / 1200.0) * 100.0
	if pct > 100 {
		return 100
	}
	if pct < 10 {
		return 10
	}
	return pct
}

// StageSelector picks a plan based on plenum capacity and zone count.
type StageSelector struct {
	zoneCount int
	targetCMH float64
	fanIDs    []model.FanID
}

func NewStageSelector(zoneCount int, targetCMH float64, fanIDs []model.FanID) *StageSelector {
	ids := make([]model.FanID, len(fanIDs))
	copy(ids, fanIDs)
	return &StageSelector{zoneCount: zoneCount, targetCMH: targetCMH, fanIDs: ids}
}

func (sel *StageSelector) Select() StagePlan {
	return BuildStagingPlan(sel.zoneCount, sel.targetCMH, sel.fanIDs)
}

func (sel *StageSelector) FanCount() int { return len(sel.fanIDs) }
