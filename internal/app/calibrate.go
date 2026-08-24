package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

// CalibrateProbe allows acceptance tests to inject feed calibration faults.
var CalibrateProbe func(ctx context.Context) error

const feedTempLimitC = 55.0

func (a *App) CalibrateFeed(ctx context.Context, tower model.TowerID, holder string) error {
	if err := a.feedLeases.Require(tower, holder, 30*time.Second); err != nil {
		return err
	}
	// Release on every exit path: a probe fault (e.g. detached temperature
	// probe) aborts calibration mid-flow, and the holder must not retain the
	// lease past this scope or the next cast start sees the mold as still held.
	defer a.feedLeases.ReleaseHolder(tower, holder)
	if CalibrateProbe != nil {
		if err := CalibrateProbe(ctx); err != nil {
			return fmt.Errorf("calibrate: %w", err)
		}
	}
	return nil
}
