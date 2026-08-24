package store

import (
	"time"

	"github.com/lacsar712/wirecast/internal/model"
)

type ScheduleStore struct {
	mem *Memory
}

func NewScheduleStore(mem *Memory) *ScheduleStore {
	return &ScheduleStore{mem: mem}
}

func (s *ScheduleStore) Save(sched model.DryingSchedule) {
	s.mem.PutSchedule(sched)
}

func (s *ScheduleStore) Load(id model.ScheduleID) (model.DryingSchedule, error) {
	sched, ok := s.mem.GetSchedule(id)
	if !ok {
		return model.DryingSchedule{}, model.Wrap("schedule", "load", model.ErrNotFound)
	}
	return sched, nil
}

func (s *ScheduleStore) SnapshotClone(id model.ScheduleID) (model.DryingSchedule, error) {
	sched, err := s.Load(id)
	if err != nil {
		return model.DryingSchedule{}, err
	}
	return sched.Clone(), nil
}

func (s *ScheduleStore) ActiveEntry(sched model.DryingSchedule, now time.Time) (model.DryingScheduleEntry, bool) {
	for _, e := range sched.Entries {
		if !now.Before(e.Start) && now.Before(e.End) {
			return e, true
		}
	}
	return model.DryingScheduleEntry{}, false
}

func (s *ScheduleStore) EntriesOverlapping(sched model.DryingSchedule, start, end time.Time) []model.DryingScheduleEntry {
	var out []model.DryingScheduleEntry
	for _, e := range sched.Entries {
		if e.End.After(start) && e.Start.Before(end) {
			out = append(out, e)
		}
	}
	return out
}
