package shelltemp

import (
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

type ProfileManager struct {
	profile model.MoistureProfile
	hold    bool
	holdAt  time.Time
	holdDur time.Duration
}

func NewProfileManager(zones []model.ZoneID, targets []float64) (*ProfileManager, error) {
	if len(zones) != len(targets) {
		return nil, model.Wrap("profile", "length_mismatch", model.ErrConflict)
	}
	return &ProfileManager{
		profile: model.MoistureProfile{Zones: zones, Targets: targets},
	}, nil
}

func (p *ProfileManager) Profile() model.MoistureProfile {
	return p.profile.Clone()
}

func (p *ProfileManager) SetProfile(profile model.MoistureProfile) error {
	if len(profile.Zones) != len(profile.Targets) {
		return model.Wrap("profile", "length_mismatch", model.ErrConflict)
	}
	p.profile = profile.Clone()
	return nil
}

func (p *ProfileManager) ArmHold(at time.Time, duration time.Duration) {
	p.hold = true
	p.holdAt = at
	p.holdDur = duration
}

func (p *ProfileManager) HoldActive(clk func() time.Time) bool {
	if !p.hold {
		return false
	}
	now := clk()
	end := p.holdAt.Add(p.holdDur)
	return !now.Before(p.holdAt) && now.Before(end)
}

func (p *ProfileManager) ReleaseHold() {
	p.hold = false
}

func (p *ProfileManager) TargetFor(zone model.ZoneID) (float64, bool) {
	for i, z := range p.profile.Zones {
		if z == zone {
			return p.profile.Targets[i], true
		}
	}
	return 0, false
}

func (p *ProfileManager) AllAtTarget(readings []model.MoistureReading, tolerance float64) bool {
	for _, r := range readings {
		target, ok := p.TargetFor(r.Zone)
		if !ok {
			continue
		}
		diff := r.Pct - target
		if diff < -tolerance || diff > tolerance {
			return false
		}
	}
	return true
}

func (p *ProfileManager) Window(start time.Time, duration time.Duration, targetPct float64) {
	for i := range p.profile.Targets {
		p.profile.Targets[i] = targetPct
	}
	p.ArmHold(start, duration)
}
