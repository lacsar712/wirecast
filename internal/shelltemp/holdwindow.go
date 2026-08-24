package shelltemp

import (
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

// HoldWindow represents one equalization hold interval in a drying profile.
type HoldWindow struct {
	Start     time.Time
	End       time.Time
	TargetPct float64
	MinHold   time.Duration
	Released  bool
	Label     string
}

func (w HoldWindow) ActiveAt(now time.Time) bool {
	if w.Released {
		return false
	}
	return !now.Before(w.Start) && now.Before(w.End)
}

func (w HoldWindow) Elapsed(now time.Time) time.Duration {
	if now.Before(w.Start) {
		return 0
	}
	end := now
	if end.After(w.End) {
		end = w.End
	}
	return end.Sub(w.Start)
}

func (w HoldWindow) MinMet(now time.Time) bool {
	return w.Elapsed(now) >= w.MinHold
}

// HoldWindowManager tracks scheduled hold windows for a drying profile.
type HoldWindowManager struct {
	windows []HoldWindow
	active  int
}

func NewHoldWindowManager() *HoldWindowManager {
	return &HoldWindowManager{active: -1}
}

func (m *HoldWindowManager) Schedule(start time.Time, duration time.Duration, targetPct float64, label string) int {
	if duration <= 0 {
		duration = time.Minute
	}
	minHold := duration / 2
	if minHold < time.Second {
		minHold = time.Second
	}
	w := HoldWindow{
		Start: start, End: start.Add(duration),
		TargetPct: targetPct, MinHold: minHold, Label: label,
	}
	m.windows = append(m.windows, w)
	idx := len(m.windows) - 1
	if m.active < 0 {
		m.active = idx
	}
	return idx
}

func (m *HoldWindowManager) Count() int { return len(m.windows) }

func (m *HoldWindowManager) ActiveWindow(now time.Time) (*HoldWindow, bool) {
	if m.active < 0 || m.active >= len(m.windows) {
		return nil, false
	}
	w := &m.windows[m.active]
	if w.ActiveAt(now) {
		return w, true
	}
	return nil, false
}

func (m *HoldWindowManager) AnyActive(now time.Time) bool {
	_, ok := m.ActiveWindow(now)
	return ok
}

func (m *HoldWindowManager) CanRelease(now time.Time, readings []model.MoistureReading, tolerance float64) bool {
	w, ok := m.ActiveWindow(now)
	if !ok {
		return true
	}
	if !w.MinMet(now) {
		return false
	}
	for _, r := range readings {
		diff := r.Pct - w.TargetPct
		if diff < -tolerance || diff > tolerance {
			return false
		}
	}
	return true
}

func (m *HoldWindowManager) ReleaseActive() {
	if m.active < 0 || m.active >= len(m.windows) {
		return
	}
	m.windows[m.active].Released = true
	m.active++
	for m.active < len(m.windows) && m.windows[m.active].Released {
		m.active++
	}
	if m.active >= len(m.windows) {
		m.active = -1
	}
}

func (m *HoldWindowManager) ReleaseAll() {
	for i := range m.windows {
		m.windows[i].Released = true
	}
	m.active = -1
}

func (m *HoldWindowManager) PendingCount(now time.Time) int {
	n := 0
	for i := range m.windows {
		if !m.windows[i].Released && m.windows[i].End.After(now) {
			n++
		}
	}
	return n
}

func (m *HoldWindowManager) NextStart(now time.Time) (time.Time, bool) {
	for i := range m.windows {
		if !m.windows[i].Released && m.windows[i].Start.After(now) {
			return m.windows[i].Start, true
		}
	}
	return time.Time{}, false
}

// HoldWindowEvaluator bridges profile manager holds with window tracking.
type HoldWindowEvaluator struct {
	profile *ProfileManager
	windows *HoldWindowManager
}

func NewHoldWindowEvaluator(profile *ProfileManager, windows *HoldWindowManager) *HoldWindowEvaluator {
	return &HoldWindowEvaluator{profile: profile, windows: windows}
}

func (e *HoldWindowEvaluator) Arm(start time.Time, duration time.Duration, targetPct float64) {
	if e.profile != nil {
		e.profile.Window(start, duration, targetPct)
	}
	if e.windows != nil {
		e.windows.Schedule(start, duration, targetPct, "profile_hold")
	}
}

func (e *HoldWindowEvaluator) HoldActive(now time.Time) bool {
	if e.profile != nil && e.profile.HoldActive(func() time.Time { return now }) {
		return true
	}
	if e.windows != nil {
		return e.windows.AnyActive(now)
	}
	return false
}

func (e *HoldWindowEvaluator) Release(now time.Time, readings []model.MoistureReading, tolerance float64) bool {
	if e.windows != nil && !e.windows.CanRelease(now, readings, tolerance) {
		return false
	}
	if e.profile != nil {
		e.profile.ReleaseHold()
	}
	if e.windows != nil {
		e.windows.ReleaseActive()
	}
	return true
}
