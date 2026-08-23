package clock

import (
	"time"

	"github.com/lacsar712/reefctl/internal/model"
)

type DefrostWindow struct {
	clk      Clock
	duration time.Duration
}

func NewDefrostWindow(clk Clock, duration time.Duration) *DefrostWindow {
	if duration <= 0 {
		duration = 5 * time.Minute
	}
	return &DefrostWindow{clk: clk, duration: duration}
}

func (w *DefrostWindow) Active(anchor time.Time) bool {
	return WindowElapsed(w.clk, anchor, w.duration)
}

func (w *DefrostWindow) Require(anchor time.Time) error {
	if w.Active(anchor) {
		return nil
	}
	return model.ErrZoneHold
}
