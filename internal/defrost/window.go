package defrost

import (
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/model"
)

type Window struct {
	Start    time.Time
	Duration time.Duration
	TargetC    float64
}

func NewWindow(start time.Time, duration time.Duration, hold float64) Window {
	return Window{Start: start, Duration: duration, TargetC: hold}
}

func (w Window) Active(clk clock.Clock) bool { return clock.WindowElapsed(clk, w.Start, w.Duration) }

func (w Window) Remaining(clk clock.Clock) time.Duration {
	end := w.Start.Add(w.Duration)
	now := clk.Now()
	if !now.Before(end) {
		return 0
	}
	return end.Sub(now)
}

func (w Window) Contains(reading model.TempReading) bool {
	end := w.Start.Add(w.Duration)
	return !reading.At.Before(w.Start) && reading.At.Before(end)
}

func (w Window) WithinTarget(reading model.TempReading) bool {
	if !w.Contains(reading) {
		return false
	}
	diff := reading.Celsius - w.TargetC
	if diff < 0 {
		diff = -diff
	}
	return diff <= 0.5
}