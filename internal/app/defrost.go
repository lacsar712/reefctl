package app

import (
	"context"

	"github.com/lacsar712/reefctl/internal/clock"
)

type DefrostPlan struct {
	PlanSteps int
}

func (a *App) ExecutePlan(ctx context.Context, plan DefrostPlan) error {
	if a.scheduler == nil {
		return nil
	}
	return a.scheduler.InstallDefrostPlanCtx(context.Background(), clock.DefrostPlan{PlanSteps: plan.PlanSteps}, "defrost-plan")
}

func (a *App) DefrostStepsDone() int {
	if a.scheduler == nil {
		return 0
	}
	return a.scheduler.DefrostStepsDone()
}
