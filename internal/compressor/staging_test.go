package compressor

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/config"
	"github.com/lacsar712/reefctl/internal/model"
)

func TestStagingInterlockDefrostBlock(t *testing.T) {
	clk := clock.NewProcessClock(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Millisecond)
	co := NewCoordinator(config.Default(), clk)
	co.Register("comp-1")
	active := true
	si := NewStagingInterlock(co, func() bool { return active }, time.Millisecond, []model.CompressorID{"comp-1"}, clk.Now)
	err := si.RequestStart(context.Background(), "comp-1")
	if err == nil {
		t.Fatal("expected defrost block")
	}
	active = false
	if err := si.RequestStart(context.Background(), "comp-1"); err != nil {
		t.Fatal(err)
	}
}
