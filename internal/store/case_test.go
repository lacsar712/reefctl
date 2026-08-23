package store

import (
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/model"
)

func TestCase(t *testing.T) {
	orig := ZoneSnapshot{
		Unit: model.UnitID("reef-unit-7"),
		DefrostSlots: []DefrostSlotSnapshot{
			{Slot: model.ZoneID("zone-00"), Duration: 15 * time.Minute},
		},
	}
	clone := CloneZoneSnapshot(orig)
	clone.DefrostSlots[0].Duration = 99 * time.Minute
	if orig.DefrostSlots[0].Duration == 99*time.Minute {
		t.Fatal("clone mutated original zone DefrostSlots backing array")
	}
}
