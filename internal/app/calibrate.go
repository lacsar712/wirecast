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
	if CalibrateProbe != nil {
		if err := CalibrateProbe(ctx); err != nil {
			a.feedLeases.ReleaseHolder(tower, holder)
			return fmt.Errorf("calibrate: %w", err)
		}
	}
	a.feedLeases.ReleaseHolder(tower, holder)
	return nil
}
