package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/config"
	"github.com/lacsar712/reefctl/internal/model"
)

func TestRunOnce(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if err := a.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if a.reeferFSM.State() != model.UnitCirculate {
		t.Fatalf("state %s", a.reeferFSM.State())
	}
}

func TestApplyScheduleSnapshot(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	now := a.clk.Now()
	a.sched.Save(model.DefrostSchedule{ID: "sch1", Entries: []model.DefrostScheduleEntry{{
		Start: now.Add(-time.Hour), End: now.Add(time.Hour), Duct: "duct-main",
		Setpoint: model.AirflowSetpoint{CubicFeetPerMinute: 8, TolerancePct: 5}, TargetCelsius: 5,
	}}})
	if err := a.ApplyScheduleSnapshot(context.Background(), "sch1"); err != nil {
		t.Fatal(err)
	}
}