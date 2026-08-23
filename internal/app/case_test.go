package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/config"
)

func TestCase(t *testing.T) {
	cfg := config.Default()
	cfg.DefrostPlanSteps = 80
	a, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = a.ExecutePlan(ctx, DefrostPlan{PlanSteps: cfg.DefrostPlanSteps})
	}()
	time.Sleep(5 * time.Millisecond)
	cancel()
	<-done
	if a.DefrostStepsDone() >= cfg.DefrostPlanSteps {
		t.Fatalf("defrost plan finished after cancel: %d steps", a.DefrostStepsDone())
	}
}
