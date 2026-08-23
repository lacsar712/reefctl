package compressor

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/config"
	"github.com/lacsar712/reefctl/internal/model"
)

func TestCoordinatorStartStop(t *testing.T) {
	cfg := config.Default()
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	co := NewCoordinator(cfg, clk)
	co.Register("comp-1")
	if err := co.Start(context.Background(), "comp-1"); err != nil {
		t.Fatal(err)
	}
	states := co.States()
	if states["comp-1"] != model.CompressorRun {
		t.Fatalf("state %s", states["comp-1"])
	}
	clk.Advance(cfg.CompressorMinRun + time.Millisecond)
	if err := co.Stop(context.Background(), "comp-1"); err != nil {
		t.Fatal(err)
	}
}

func TestCoordinatorMissingUnit(t *testing.T) {
	cfg := config.Default()
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	co := NewCoordinator(cfg, clk)
	if err := co.Start(context.Background(), "missing"); err == nil {
		t.Fatal("expected error")
	}
}
