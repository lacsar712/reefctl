package defrost

import (
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
)

func TestHeaterControllerEngageRelease(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	h := NewHeaterController(clk, time.Minute)
	if err := h.Engage("duct-evap"); err != nil {
		t.Fatal(err)
	}
	if !h.Active("duct-evap") {
		t.Fatal("expected active")
	}
	h.Release("duct-evap")
	if h.Active("duct-evap") {
		t.Fatal("expected released")
	}
}

func TestSchedulePlannerWindow(t *testing.T) {
	p := NewSchedulePlanner(time.Minute, 10*time.Minute)
	now := time.Unix(0, 0)
	w := p.PlanWindow(now, 3.0, 5.0)
	if w.Duration < time.Minute {
		t.Fatal("duration too short")
	}
	if w.TargetC != 5.0 {
		t.Fatal("target")
	}
}

func TestSchedulePlannerOverlap(t *testing.T) {
	p := NewSchedulePlanner(time.Minute, 10*time.Minute)
	now := time.Unix(0, 0)
	existing := []Window{NewWindow(now, 5*time.Minute, 4.0)}
	candidate := NewWindow(now.Add(2*time.Minute), 3*time.Minute, 4.0)
	if !p.Overlaps(existing, candidate) {
		t.Fatal("expected overlap")
	}
}
