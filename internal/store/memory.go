package store

import (
	"sync"

	"github.com/lacsar712/wirecast/internal/model"
)

type Memory struct {
	mu        sync.RWMutex
	towers    map[model.TowerID]model.TowerSnapshot
	schedules map[model.ScheduleID]model.DryingSchedule
	profiles  map[model.TowerID]model.MoistureProfile
}

func NewMemory() *Memory {
	return &Memory{
		towers:    make(map[model.TowerID]model.TowerSnapshot),
		schedules: make(map[model.ScheduleID]model.DryingSchedule),
		profiles:  make(map[model.TowerID]model.MoistureProfile),
	}
}

func (m *Memory) PutTower(snap model.TowerSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.towers[snap.ID] = snap
}

func (m *Memory) GetTower(id model.TowerID) (model.TowerSnapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.towers[id]
	return s, ok
}

func (m *Memory) PutSchedule(s model.DryingSchedule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.schedules[s.ID] = s
}

func (m *Memory) GetSchedule(id model.ScheduleID) (model.DryingSchedule, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.schedules[id]
	return s, ok
}

func (m *Memory) PutProfile(tower model.TowerID, profile model.MoistureProfile) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.profiles[tower] = profile
}

func (m *Memory) GetProfile(tower model.TowerID) (model.MoistureProfile, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.profiles[tower]
	return p, ok
}

func (m *Memory) ListTowers() []model.TowerSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.TowerSnapshot, 0, len(m.towers))
	for _, v := range m.towers {
		out = append(out, v)
	}
	return out
}
