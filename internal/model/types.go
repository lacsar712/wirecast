package model

import "time"

type TowerState string

const (
	TowerIdle      TowerState = "idle"
	TowerPreheat   TowerState = "preheat"
	TowerDrying    TowerState = "drying"
	TowerEqualize  TowerState = "equalize"
	TowerCool      TowerState = "cool"
	TowerFault     TowerState = "fault"
	TowerShutdown  TowerState = "shutdown"
)

type FanState string

const (
	FanOff     FanState = "off"
	FanRamp    FanState = "ramp"
	FanRun     FanState = "run"
	FanCoast   FanState = "coast"
	FanTrip    FanState = "trip"
)

type DamperPosition string

const (
	DamperClosed    DamperPosition = "closed"
	DamperOpen      DamperPosition = "open"
	DamperThrottled DamperPosition = "throttled"
)

type AirflowSetpoint struct {
	CubicMetersPerHour float64
	TolerancePct       float64
}

func (a AirflowSetpoint) Within(actual float64) bool {
	if a.CubicMetersPerHour <= 0 {
		return actual <= 0
	}
	lo := a.CubicMetersPerHour * (1 - a.TolerancePct/100)
	hi := a.CubicMetersPerHour * (1 + a.TolerancePct/100)
	return actual >= lo && actual <= hi
}

type MoistureReading struct {
	Sensor  SensorID
	Zone    ZoneID
	Pct     float64
	At      time.Time
}

type ZoneAssignment struct {
	Zone       ZoneID
	Plenum     PlenumID
	Setpoint   AirflowSetpoint
	Enabled    bool
	LastFlow   float64
	TargetMoist float64
	ActualMoist float64
}

type TowerSnapshot struct {
	ID        TowerID
	State     TowerState
	Zones     []ZoneAssignment
	Fans      []FanID
	UpdatedAt time.Time
}

type DryingScheduleEntry struct {
	ID              ScheduleID
	Plenum          PlenumID
	Start           time.Time
	End             time.Time
	Setpoint        AirflowSetpoint
	TargetMoistPct  float64
	EqualizeMinutes int
}

type DryingSchedule struct {
	ID      ScheduleID
	Entries []DryingScheduleEntry
	Version int64
}

func (s DryingSchedule) Clone() DryingSchedule {
	out := DryingSchedule{ID: s.ID, Version: s.Version}
	if len(s.Entries) == 0 {
		return out
	}
	out.Entries = make([]DryingScheduleEntry, len(s.Entries))
	copy(out.Entries, s.Entries)
	return out
}

type AlarmEvent struct {
	Code     AlarmCode
	Message  string
	Tower    TowerID
	RaisedAt time.Time
	Severity int
}

type PlenumRoute struct {
	From     PlenumID
	To       PlenumID
	Damper   DamperID
	Priority int
}

type MoistureProfile struct {
	Zones   []ZoneID
	Targets []float64
}

func (p MoistureProfile) Clone() MoistureProfile {
	out := MoistureProfile{}
	if len(p.Zones) > 0 {
		out.Zones = make([]ZoneID, len(p.Zones))
		copy(out.Zones, p.Zones)
	}
	if len(p.Targets) > 0 {
		out.Targets = make([]float64, len(p.Targets))
		copy(out.Targets, p.Targets)
	}
	return out
}

type GradientSample struct {
	Zone    ZoneID
	Reading float64
	At      time.Time
}
