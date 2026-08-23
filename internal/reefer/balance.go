package reefer

import (
	"context"
	"fmt"

	"github.com/lacsar712/reefctl/internal/model"
)

type DuctDemand struct {
	Duct     model.DuctID
	SharePct float64
	Capacity float64
}

type BalancePlan struct {
	TotalCFM  float64
	Tolerance float64
	Entries   []DuctDemand
}

func BuildBalancePlan(registry *DuctRegistry, parts *PartitionTable, totalCFM, tolerancePct float64) (BalancePlan, error) {
	if parts == nil {
		return BalancePlan{}, model.Wrap("balance", "no_partitions", model.ErrNotFound)
	}
	plan := BalancePlan{TotalCFM: totalCFM, Tolerance: tolerancePct}
	seen := make(map[model.DuctID]struct{})
	for _, p := range parts.Partitions() {
		if _, dup := seen[p.Duct]; dup {
			continue
		}
		seen[p.Duct] = struct{}{}
		capacity := totalCFM
		if m, ok := registry.Get(p.Duct); ok && m.Capacity > 0 {
			capacity = m.Capacity
		}
		plan.Entries = append(plan.Entries, DuctDemand{
			Duct:     p.Duct,
			SharePct: p.SharePct,
			Capacity: capacity,
		})
	}
	if len(plan.Entries) == 0 {
		return BalancePlan{}, model.Wrap("balance", "empty", model.ErrNotFound)
	}
	return plan, nil
}

func (plan BalancePlan) Setpoints() map[model.DuctID]model.AirflowSetpoint {
	out := make(map[model.DuctID]model.AirflowSetpoint, len(plan.Entries))
	for _, e := range plan.Entries {
		cfm := plan.TotalCFM * e.SharePct / 100
		if cfm > e.Capacity {
			cfm = e.Capacity
		}
		out[e.Duct] = model.AirflowSetpoint{CubicFeetPerMinute: cfm, TolerancePct: plan.Tolerance}
	}
	return out
}

type DuctBalancer struct {
	registry *DuctRegistry
	flow     map[model.DuctID]*AirflowController
}

func NewDuctBalancer(registry *DuctRegistry) *DuctBalancer {
	return &DuctBalancer{registry: registry, flow: make(map[model.DuctID]*AirflowController)}
}

func (b *DuctBalancer) ApplyPlan(plan BalancePlan) {
	for duct, sp := range plan.Setpoints() {
		if fc, ok := b.flow[duct]; ok {
			fc.SetSetpoint(sp)
		} else {
			b.flow[duct] = NewAirflowController(sp)
		}
		if m, ok := b.registry.Get(duct); ok {
			m.SetFlow(sp.CubicFeetPerMinute)
		}
	}
}

func (b *DuctBalancer) Observe(duct model.DuctID, cfm float64) error {
	fc, ok := b.flow[duct]
	if !ok {
		return model.Wrap("balance", string(duct), model.ErrNotFound)
	}
	fc.Observe(cfm)
	return nil
}

func (b *DuctBalancer) Trim(ctx context.Context, overDuct, underDuct model.DuctID, shiftPct float64) error {
	over, ok := b.flow[overDuct]
	if !ok {
		return model.Wrap("balance", "over", model.ErrNotFound)
	}
	under, ok := b.flow[underDuct]
	if !ok {
		return model.Wrap("balance", "under", model.ErrNotFound)
	}
	select {
	case <-ctx.Done():
		return model.Wrap("balance", "canceled", context.Cause(ctx))
	default:
	}
	overSP := over.Setpoint()
	underSP := under.Setpoint()
	delta := overSP.CubicFeetPerMinute * shiftPct / 100
	if delta <= 0 {
		return nil
	}
	over.SetSetpoint(model.AirflowSetpoint{
		CubicFeetPerMinute: overSP.CubicFeetPerMinute - delta,
		TolerancePct:       overSP.TolerancePct,
	})
	under.SetSetpoint(model.AirflowSetpoint{
		CubicFeetPerMinute: underSP.CubicFeetPerMinute + delta,
		TolerancePct:       underSP.TolerancePct,
	})
	return nil
}

func (b *DuctBalancer) ValidateAll(ctx context.Context) error {
	for duct, fc := range b.flow {
		if err := fc.Validate(ctx); err != nil {
			return model.Wrap("balance", string(duct), err)
		}
	}
	return nil
}

func (b *DuctBalancer) Summary() string {
	parts := make([]string, 0, len(b.flow))
	for duct, fc := range b.flow {
		sp := fc.Setpoint()
		parts = append(parts, fmt.Sprintf("%s=%.1f/%.1f", duct, fc.Actual(), sp.CubicFeetPerMinute))
	}
	return fmt.Sprintf("duct_balance {%s}", stringsJoin(parts, ","))
}
