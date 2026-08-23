package store

import (
	"sync"

	"github.com/lacsar712/reefctl/internal/model"
)

type Memory struct {
	mu        sync.RWMutex
	racks     map[model.UnitID]model.UnitSnapshot
	schedules map[model.ScheduleID]model.DefrostSchedule
}

func NewMemory() *Memory {
	return &Memory{
		racks:     make(map[model.UnitID]model.UnitSnapshot),
		schedules: make(map[model.ScheduleID]model.DefrostSchedule),
	}
}

func (m *Memory) PutUnit(snap model.UnitSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.racks[snap.ID] = snap
}

func (m *Memory) GetUnit(id model.UnitID) (model.UnitSnapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.racks[id]
	return s, ok
}

func (m *Memory) PutSchedule(s model.DefrostSchedule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.schedules[s.ID] = s
}

func (m *Memory) GetSchedule(id model.ScheduleID) (model.DefrostSchedule, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.schedules[id]
	return s, ok
}

func (m *Memory) ListUnits() []model.UnitSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.UnitSnapshot, 0, len(m.racks))
	for _, v := range m.racks {
		out = append(out, v)
	}
	return out
}