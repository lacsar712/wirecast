package fsm

import (
	"fmt"

	"github.com/lacsar712/wirecast/internal/model"
)

type Transition struct {
	From  model.TowerState
	To    model.TowerState
	Event string
}

var towerTransitions = []Transition{
	{model.TowerIdle, model.TowerPreheat, "preheat"},
	{model.TowerPreheat, model.TowerDrying, "airflow_ok"},
	{model.TowerDrying, model.TowerEqualize, "moisture_hold"},
	{model.TowerEqualize, model.TowerDrying, "release_hold"},
	{model.TowerDrying, model.TowerCool, "target_reached"},
	{model.TowerCool, model.TowerIdle, "cool_complete"},
	{model.TowerPreheat, model.TowerFault, "fault"},
	{model.TowerDrying, model.TowerFault, "fault"},
	{model.TowerEqualize, model.TowerFault, "fault"},
	{model.TowerCool, model.TowerFault, "fault"},
	{model.TowerFault, model.TowerShutdown, "shutdown"},
	{model.TowerIdle, model.TowerShutdown, "shutdown"},
}

func AllowedTower(from model.TowerState, event string) (model.TowerState, bool) {
	for _, tr := range towerTransitions {
		if tr.From == from && tr.Event == event {
			return tr.To, true
		}
	}
	return from, false
}

func MustTower(from model.TowerState, event string) (model.TowerState, error) {
	to, ok := AllowedTower(from, event)
	if !ok {
		return from, model.Wrap("tower_fsm", "illegal_transition", fmt.Errorf("%s -> %s", from, event))
	}
	return to, nil
}

var fanTransitions = []struct {
	from, to model.FanState
	event    string
}{
	{model.FanOff, model.FanRamp, "start"},
	{model.FanRamp, model.FanRun, "ramped"},
	{model.FanRun, model.FanCoast, "stop"},
	{model.FanCoast, model.FanOff, "coast_done"},
	{model.FanRun, model.FanTrip, "trip"},
	{model.FanRamp, model.FanTrip, "trip"},
}

func AllowedFan(from model.FanState, event string) (model.FanState, bool) {
	for _, tr := range fanTransitions {
		if tr.from == from && tr.event == event {
			return tr.to, true
		}
	}
	return from, false
}

func MustFan(from model.FanState, event string) (model.FanState, error) {
	to, ok := AllowedFan(from, event)
	if !ok {
		return from, model.Wrap("fan_fsm", "illegal_transition", fmt.Errorf("%s -> %s", from, event))
	}
	return to, nil
}
