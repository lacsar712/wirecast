package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/lacsar712/wirecast/internal/config"
	"github.com/lacsar712/wirecast/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	tower := model.TowerID(a.cfg.TowerID)
	CalibrateProbe = func(ctx context.Context) error {
		return fmt.Errorf("feed probe fault")
	}
	if err := a.CalibrateFeed(context.Background(), tower, "operator-a"); err == nil {
		t.Fatal("expected calibration probe failure")
	}
	CalibrateProbe = nil
	if err := a.CalibrateFeed(context.Background(), tower, "operator-b"); err != nil {
		t.Fatalf("second calibration blocked by leaked feed lease: %v", err)
	}
}
