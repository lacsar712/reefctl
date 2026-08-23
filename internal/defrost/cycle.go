package defrost

import (
	"context"
	"errors"
	"sync"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/model"
)

type CycleController struct {
	mu     sync.Mutex
	clk    clock.Clock
	window Window
	active bool
}

func NewCycleController(clk clock.Clock) *CycleController { return &CycleController{clk: clk} }

func (h *CycleController) Arm(w Window) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.window = w
	h.active = true
}

func (h *CycleController) Release() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.active = false
}

func (h *CycleController) Active() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.active && h.window.Active(h.clk)
}

func (h *CycleController) WaitStable(ctx context.Context, readings <-chan model.TempReading) error {
	for {
		select {
		case <-ctx.Done():
			return errors.Join(model.ErrContextCanceled, context.Cause(ctx))
		case r, ok := <-readings:
			if !ok {
				return model.ErrDefrostActive
			}
			h.mu.Lock()
			w := h.window
			act := h.active
			h.mu.Unlock()
			if !act {
				return nil
			}
			if w.WithinTarget(r) && !w.Active(h.clk) {
				return nil
			}
		}
	}
}