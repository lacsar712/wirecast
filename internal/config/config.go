package config

import "time"

type Config struct {
	TowerID              string
	ZoneCount            int
	DefaultAirflowCMH    float64
	AirflowTolerancePct  float64
	EqualizeHoldMinutes  int
	FanMinRun            time.Duration
	FanCoast             time.Duration
	PlenumPrimeSec       int
	AlarmBufferSize      int
	ProcessTickMs        int
	TargetMoistPct       float64
	MaxGradientDeltaPct  float64
	DryingRampMinutes    int
	CastRampSteps         int
	SegmentVentSteps     int
	SolidWindowMinutes int
}

func Default() Config {
	return Config{
		TowerID: "tower-a1", ZoneCount: 6, DefaultAirflowCMH: 850.0, AirflowTolerancePct: 5,
		EqualizeHoldMinutes: 1, FanMinRun: time.Millisecond, FanCoast: time.Second,
		PlenumPrimeSec: 5, AlarmBufferSize: 64, ProcessTickMs: 10,
		TargetMoistPct: 14.0, MaxGradientDeltaPct: 2.5, DryingRampMinutes: 2,
		CastRampSteps: 40, SegmentVentSteps: 60, SolidWindowMinutes: 2,
	}
}

func (c Config) Validate() error {
	if c.ZoneCount <= 0 {
		return errConfig("zone_count must be positive")
	}
	if c.DefaultAirflowCMH < 0 {
		return errConfig("default_airflow_cmh invalid")
	}
	if c.MaxGradientDeltaPct <= 0 {
		return errConfig("max_gradient_delta_pct must be positive")
	}
	return nil
}

func (c Config) ProcessTick() time.Duration {
	if c.ProcessTickMs <= 0 {
		return 100 * time.Millisecond
	}
	return time.Duration(c.ProcessTickMs) * time.Millisecond
}
