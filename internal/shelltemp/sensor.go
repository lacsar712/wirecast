package shelltemp

import (
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

type Sensor struct {
	ID    model.SensorID
	Zone  model.ZoneID
	value float64
	at    time.Time
}

func NewSensor(id model.SensorID, zone model.ZoneID) *Sensor {
	return &Sensor{ID: id, Zone: zone}
}

func (s *Sensor) Observe(pct float64, at time.Time) model.MoistureReading {
	s.value = pct
	s.at = at
	return model.MoistureReading{Sensor: s.ID, Zone: s.Zone, Pct: pct, At: at}
}

func (s *Sensor) Last() (float64, time.Time) {
	return s.value, s.at
}

type SensorBank struct {
	sensors map[model.ZoneID]*Sensor
}

func NewSensorBank() *SensorBank {
	return &SensorBank{sensors: make(map[model.ZoneID]*Sensor)}
}

func (b *SensorBank) Register(s *Sensor) {
	b.sensors[s.Zone] = s
}

func (b *SensorBank) ObserveZone(zone model.ZoneID, pct float64, at time.Time) (model.MoistureReading, error) {
	s, ok := b.sensors[zone]
	if !ok {
		return model.MoistureReading{}, model.Wrap("sensor_bank", "zone", model.ErrNotFound)
	}
	return s.Observe(pct, at), nil
}

func (b *SensorBank) Readings() []model.MoistureReading {
	out := make([]model.MoistureReading, 0, len(b.sensors))
	for _, s := range b.sensors {
		pct, at := s.Last()
		out = append(out, model.MoistureReading{Sensor: s.ID, Zone: s.Zone, Pct: pct, At: at})
	}
	return out
}

func (b *SensorBank) ZoneCount() int { return len(b.sensors) }
