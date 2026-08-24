package clock

import (
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

type SolidWindow struct {
	clk      Clock
	duration time.Duration
}

func NewSolidWindow(clk Clock, duration time.Duration) *SolidWindow {
	if duration <= 0 {
		duration = 2 * time.Minute
	}
	return &SolidWindow{clk: clk, duration: duration}
}

func (w *SolidWindow) Active(anchor time.Time) bool {
	return time.Since(anchor) < w.duration
}

func (w *SolidWindow) Require(anchor time.Time) error {
	if w.Active(anchor) {
		return nil
	}
	return model.ErrSolidHold
}
