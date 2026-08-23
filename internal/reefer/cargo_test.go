package reefer

import (
	"testing"

	"github.com/lacsar712/reefctl/internal/model"
)

func TestCargoMonitorViolations(t *testing.T) {
	m := NewCargoMonitor()
	zone := model.ZoneID("reef-unit-7-zone-00")
	m.SetProfile(CargoProfile{ZoneID: zone, MinCelsius: -2, MaxCelsius: 4, Sensitivity: 1})
	m.Record(zone, 8.0)
	v := m.Violations()
	if len(v) != 1 || v[0] != zone {
		t.Fatalf("violations %v", v)
	}
	id, dev, ok := m.WorstDeviation()
	if !ok || id != zone || dev != 4.0 {
		t.Fatalf("worst %s %.1f %v", id, dev, ok)
	}
}
