package model

import (
	"fmt"
	"strings"
)

type TowerID string
type ZoneID string
type FanID string
type DamperID string
type SensorID string
type ScheduleID string
type AlarmCode string
type PlenumID string

func (id TowerID) String() string     { return string(id) }
func (id ZoneID) String() string       { return string(id) }
func (id FanID) String() string         { return string(id) }
func (id DamperID) String() string      { return string(id) }
func (id SensorID) String() string      { return string(id) }
func (id ScheduleID) String() string    { return string(id) }
func (id AlarmCode) String() string     { return string(id) }
func (id PlenumID) String() string      { return string(id) }

func ParseTowerID(raw string) (TowerID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return TowerID(raw), nil
}

func ParseZoneID(tower TowerID, index int) (ZoneID, error) {
	if tower == "" || index < 0 {
		return "", ErrInvalidID
	}
	return ZoneID(fmt.Sprintf("%s-zone-%02d", tower, index)), nil
}

func ParseFanID(raw string) (FanID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return FanID(raw), nil
}

func ParseDamperID(raw string) (DamperID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return DamperID(raw), nil
}

func ParseSensorID(raw string) (SensorID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return SensorID(raw), nil
}

func ParseScheduleID(raw string) (ScheduleID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return ScheduleID(raw), nil
}

func ParsePlenumID(raw string) (PlenumID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return PlenumID(raw), nil
}
