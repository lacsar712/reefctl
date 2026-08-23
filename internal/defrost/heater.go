package defrost

import (
	"context"
	"sync"
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/model"
)

// HeaterController drives evaporator heater elements during defrost.
type HeaterController struct {
	mu      sync.Mutex
	clk     clock.Clock
	active  map[model.DuctID]time.Time
	maxDuty time.Duration
}

func NewHeaterController(clk clock.Clock, maxDuty time.Duration) *HeaterController {
	if maxDuty <= 0 {
		maxDuty = 30 * time.Minute
	}
	return &HeaterController{clk: clk, active: make(map[model.DuctID]time.Time), maxDuty: maxDuty}
}

func (h *HeaterController) Engage(id model.DuctID) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if start, ok := h.active[id]; ok {
		if h.clk.Now().Sub(start) > h.maxDuty {
			return model.Wrap("heater", "duty_limit", model.ErrConflict)
		}
		return nil
	}
	h.active[id] = h.clk.Now()
	return nil
}

func (h *HeaterController) Release(id model.DuctID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.active, id)
}

func (h *HeaterController) Active(id model.DuctID) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, ok := h.active[id]
	return ok
}

func (h *HeaterController) ActiveDucts() []model.DuctID {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]model.DuctID, 0, len(h.active))
	for id := range h.active {
		out = append(out, id)
	}
	return out
}

func (h *HeaterController) RunUntilClear(ctx context.Context, id model.DuctID, frostLevel float64, meltPerTick float64) error {
	for frostLevel > 0 {
		select {
		case <-ctx.Done():
			return model.Wrap("heater", "canceled", context.Cause(ctx))
		default:
		}
		if err := h.Engage(id); err != nil {
			return err
		}
		frostLevel -= meltPerTick
		if pc, ok := h.clk.(*clock.ProcessClock); ok {
			pc.Step()
		}
	}
	h.Release(id)
	return nil
}

// SchedulePlanner builds defrost windows from frost accumulation readings.
type SchedulePlanner struct {
	minInterval time.Duration
	maxDuration time.Duration
}

func NewSchedulePlanner(minInterval, maxDuration time.Duration) *SchedulePlanner {
	return &SchedulePlanner{minInterval: minInterval, maxDuration: maxDuration}
}

func (p *SchedulePlanner) PlanWindow(now time.Time, frostIndex float64, targetC float64) Window {
	duration := p.maxDuration
	if frostIndex < 1 {
		duration = p.minInterval
	} else if frostIndex > 5 {
		duration = p.maxDuration
	} else {
		duration = p.minInterval + time.Duration(frostIndex)*time.Minute
	}
	if duration > p.maxDuration {
		duration = p.maxDuration
	}
	return NewWindow(now, duration, targetC)
}

func (p *SchedulePlanner) Overlaps(existing []Window, candidate Window) bool {
	for _, w := range existing {
		end := w.Start.Add(w.Duration)
		candEnd := candidate.Start.Add(candidate.Duration)
		if candidate.Start.Before(end) && candEnd.After(w.Start) {
			return true
		}
	}
	return false
}
