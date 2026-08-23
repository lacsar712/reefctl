package reefer

import (
	"context"
	"fmt"

	"github.com/lacsar712/reefctl/internal/model"
)

type CoordinationPlan struct {
	Compressors []model.CompressorID
	Ducts       []model.DuctID
	Setpoints   map[model.DuctID]model.AirflowSetpoint
}

func BuildCoordinationPlan(registry *DuctRegistry, defaultSP model.AirflowSetpoint) CoordinationPlan {
	plan := CoordinationPlan{Setpoints: make(map[model.DuctID]model.AirflowSetpoint)}
	for _, m := range registry.All() {
		plan.Ducts = append(plan.Ducts, m.ID)
		plan.Setpoints[m.ID] = defaultSP
	}
	return plan
}

func (plan CoordinationPlan) Validate() error {
	if len(plan.Ducts) == 0 {
		return model.Wrap("coordination", "empty", model.ErrNotFound)
	}
	for _, id := range plan.Ducts {
		sp, ok := plan.Setpoints[id]
		if !ok {
			return model.Wrap("coordination", string(id), model.ErrAirflowSetpoint)
		}
		if sp.CubicFeetPerMinute < 0 {
			return model.Wrap("coordination", string(id), fmt.Errorf("negative setpoint"))
		}
	}
	return nil
}

func (p *ReeferUnit) ExecutePlan(ctx context.Context, plan CoordinationPlan) error {
	if err := plan.Validate(); err != nil {
		return err
	}
	for _, duct := range plan.Ducts {
		if err := p.PrimeDuct(ctx, duct); err != nil {
			return err
		}
		p.BindFlow(duct, plan.Setpoints[duct])
	}
	for _, id := range plan.Compressors {
		if err := p.coordinator.Start(ctx, id); err != nil {
			return err
		}
	}
	return p.ValidateFlows(ctx)
}
