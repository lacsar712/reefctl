package defrost

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/model"
)

func TestPhaseControllerLifecycle(t *testing.T) {
	clk := clock.NewProcessClock(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Millisecond)
	heater := NewHeaterController(clk, time.Minute)
	pc := NewPhaseController(clk, heater, "duct-main")
	win := NewWindow(clk.Now(), time.Minute, 6.0)
	pc.Begin(win)
	if pc.Phase() != PhasePreheat {
		t.Fatalf("phase %s", pc.Phase())
	}
	reading := model.TempReading{Sensor: "evap", Celsius: 5.5, At: clk.Now()}
	if err := pc.Advance(context.Background(), 2.0, reading); err != nil {
		t.Fatal(err)
	}
	if pc.Phase() != PhaseMelt {
		t.Fatalf("expected melt got %s", pc.Phase())
	}
	if err := pc.Advance(context.Background(), 0, reading); err != nil {
		t.Fatal(err)
	}
	if pc.Phase() != PhaseDrip {
		t.Fatalf("expected drip got %s", pc.Phase())
	}
	reading.Celsius = 3.5
	if err := pc.Advance(context.Background(), 0, reading); err != nil {
		t.Fatal(err)
	}
	if pc.Phase() != PhaseSettle {
		t.Fatalf("expected settle got %s", pc.Phase())
	}
}
