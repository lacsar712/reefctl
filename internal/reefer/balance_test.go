package reefer

import (
	"context"
	"testing"

	"github.com/lacsar712/reefctl/internal/model"
)

func TestDuctBalancerApplyPlan(t *testing.T) {
	reg := NewDuctRegistry()
	reg.Add(&Duct{ID: "duct-main", Capacity: 50})
	unit, _ := model.ParseUnitID("reef-unit-7")
	zones, _ := NewZoneTable(unit, 2, "duct-main")
	_ = zones.Enable(0, true)
	pt := NewPartitionTable(unit)
	_ = pt.BuildFromZones(zones, nil)
	plan, err := BuildBalancePlan(reg, pt, 20, 5)
	if err != nil {
		t.Fatal(err)
	}
	bal := NewDuctBalancer(reg)
	bal.ApplyPlan(plan)
	_ = bal.Observe("duct-main", 20)
	if err := bal.ValidateAll(context.Background()); err != nil {
		t.Fatal(err)
	}
}
