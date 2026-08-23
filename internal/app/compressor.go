package app

import (
	"context"
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
)

type CompressorStart struct {
	clk      clock.Clock
	tick     time.Duration
	steps    int
	freqPct  float64
}

func NewCompressorStart(clk clock.Clock, tick time.Duration, steps int) *CompressorStart {
	if steps <= 0 {
		steps = 40
	}
	return &CompressorStart{clk: clk, tick: tick, steps: steps}
}

func (c *CompressorStart) FreqPct() float64 { return c.freqPct }

func (c *CompressorStart) Start(ctx context.Context) error {
	step := 100.0 / float64(c.steps)
	cur := 0.0
	for cur < 100 {
		cur += step
		if cur > 100 {
			cur = 100
		}
		c.freqPct = cur
		if pc, ok := c.clk.(*clock.ProcessClock); ok {
			pc.Step()
		}
		time.Sleep(2 * time.Millisecond)
	}
	return nil
}

func (a *App) RunCompressorStart(ctx context.Context) error {
	return a.compressorStart.Start(ctx)
}

func (a *App) CompressorFreqPct() float64 { return a.compressorStart.FreqPct() }
