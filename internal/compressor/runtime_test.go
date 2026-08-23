package compressor

import (
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/model"
)

func TestRuntimeStatsAccumulate(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Second)
	stats := NewRuntimeStats(clk)
	stats.OnStart("comp-1")
	clk.Advance(5 * time.Second)
	stats.OnStop("comp-1")
	if stats.TotalRunSeconds("comp-1") < 5 {
		t.Fatalf("run seconds %f", stats.TotalRunSeconds("comp-1"))
	}
	if stats.StartCount("comp-1") != 1 {
		t.Fatal("start count")
	}
}

func TestLoadBalancerPick(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Second)
	stats := NewRuntimeStats(clk)
	stats.OnStart("comp-a")
	clk.Advance(10 * time.Second)
	stats.OnStop("comp-a")
	lb := NewLoadBalancer(stats)
	id, err := lb.Pick([]model.CompressorID{"comp-a", "comp-b"})
	if err != nil {
		t.Fatal(err)
	}
	if id != "comp-b" {
		t.Fatalf("picked %s", id)
	}
}
