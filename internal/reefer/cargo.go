package reefer

import (
	"fmt"
	"sync"

	"github.com/lacsar712/reefctl/internal/model"
)

// CargoProfile describes per-zone cargo temperature requirements.
type CargoProfile struct {
	ZoneID      model.ZoneID
	MinCelsius  float64
	MaxCelsius  float64
	Sensitivity float64
}

func (p CargoProfile) WithinRange(celsius float64) bool {
	return celsius >= p.MinCelsius && celsius <= p.MaxCelsius
}

func (p CargoProfile) Deviation(celsius float64) float64 {
	if celsius < p.MinCelsius {
		return p.MinCelsius - celsius
	}
	if celsius > p.MaxCelsius {
		return celsius - p.MaxCelsius
	}
	return 0
}

// CargoMonitor tracks zone cargo readings against profiles.
type CargoMonitor struct {
	mu       sync.RWMutex
	profiles map[model.ZoneID]CargoProfile
	readings map[model.ZoneID]float64
}

func NewCargoMonitor() *CargoMonitor {
	return &CargoMonitor{
		profiles: make(map[model.ZoneID]CargoProfile),
		readings: make(map[model.ZoneID]float64),
	}
}

func (m *CargoMonitor) SetProfile(p CargoProfile) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.profiles[p.ZoneID] = p
}

func (m *CargoMonitor) Record(zone model.ZoneID, celsius float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readings[zone] = celsius
}

func (m *CargoMonitor) Violations() []model.ZoneID {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []model.ZoneID
	for id, temp := range m.readings {
		p, ok := m.profiles[id]
		if !ok {
			continue
		}
		if !p.WithinRange(temp) {
			out = append(out, id)
		}
	}
	return out
}

func (m *CargoMonitor) WorstDeviation() (model.ZoneID, float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var best model.ZoneID
	var maxDev float64
	found := false
	for id, temp := range m.readings {
		p, ok := m.profiles[id]
		if !ok {
			continue
		}
		dev := p.Deviation(temp)
		if dev > maxDev {
			maxDev = dev
			best = id
			found = true
		}
	}
	return best, maxDev, found
}

func (m *CargoMonitor) Summary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	parts := make([]string, 0, len(m.readings))
	for id, temp := range m.readings {
		parts = append(parts, fmt.Sprintf("%s=%.1fC", id, temp))
	}
	return fmt.Sprintf("cargo {%s}", stringsJoin(parts, ","))
}
