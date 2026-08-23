package clock

import (
	"context"
	"time"

	"github.com/lacsar712/reefctl/internal/model"
)

type UnitScheduler struct {
	clk              ProcessClock
	defrostStepsDone int
}

func NewUnitScheduler(clk ProcessClock) *UnitScheduler {
	return &UnitScheduler{clk: clk}
}

func (s *UnitScheduler) DefrostStepsDone() int { return s.defrostStepsDone }

func (s *UnitScheduler) InstallDefrostPlan(settings DefrostPlan, planID string) {
	_ = s.InstallDefrostPlanCtx(context.Background(), settings, planID)
}

func (s *UnitScheduler) InstallDefrostPlanCtx(ctx context.Context, settings DefrostPlan, planID string) error {
	steps := settings.PlanSteps
	if steps <= 0 {
		steps = 60
	}
	for i := 0; i < steps; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.defrostStepsDone = i + 1
		s.clk.Step()
		time.Sleep(2 * time.Millisecond)
	}
	_ = planID
	_ = settings.Entries
	return nil
}

type DefrostPlan struct {
	PlanSteps int
	Entries   []model.DefrostScheduleEntry
}
