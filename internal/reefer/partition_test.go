package reefer

import (
	"testing"

	"github.com/lacsar712/reefctl/internal/model"
)

func TestPartitionTableBuild(t *testing.T) {
	unit, _ := model.ParseUnitID("reef-unit-7")
	zones, err := NewZoneTable(unit, 4, "duct-main")
	if err != nil {
		t.Fatal(err)
	}
	_ = zones.Enable(0, true)
	_ = zones.Enable(1, true)
	pt := NewPartitionTable(unit)
	if err := pt.BuildFromZones(zones, NewCargoMonitor()); err != nil {
		t.Fatal(err)
	}
	parts := pt.Partitions()
	if len(parts) != 1 {
		t.Fatalf("partitions %d", len(parts))
	}
	if parts[0].SharePct != 100 {
		t.Fatalf("share %.1f", parts[0].SharePct)
	}
}
