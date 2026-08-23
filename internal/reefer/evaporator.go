package reefer

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/model"
)

// Evaporator models the evaporator coil section tied to a duct path.
type Evaporator struct {
	ID         model.DuctID
	FrostIndex float64
	HeaterOn   bool
	clk        clock.Clock
}

func NewEvaporator(id model.DuctID, clk clock.Clock) *Evaporator {
	return &Evaporator{ID: id, clk: clk}
}

func (e *Evaporator) AccumulateFrost(delta float64) {
	if delta > 0 {
		e.FrostIndex += delta
	}
}

func (e *Evaporator) DefrostProgress(rate float64) {
	if e.FrostIndex <= 0 {
		return
	}
	e.FrostIndex -= rate
	if e.FrostIndex < 0 {
		e.FrostIndex = 0
	}
}

func (e *Evaporator) NeedsDefrost(threshold float64) bool {
	return e.FrostIndex >= threshold
}

func (e *Evaporator) EnableHeater(on bool) {
	e.HeaterOn = on
}

type EvaporatorBank struct {
	mu    sync.RWMutex
	clk   clock.Clock
	coils map[model.DuctID]*Evaporator
}

func NewEvaporatorBank(clk clock.Clock) *EvaporatorBank {
	return &EvaporatorBank{clk: clk, coils: make(map[model.DuctID]*Evaporator)}
}

func (b *EvaporatorBank) Register(id model.DuctID) *Evaporator {
	b.mu.Lock()
	defer b.mu.Unlock()
	if ev, ok := b.coils[id]; ok {
		return ev
	}
	ev := NewEvaporator(id, b.clk)
	b.coils[id] = ev
	return ev
}

func (b *EvaporatorBank) Get(id model.DuctID) (*Evaporator, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	ev, ok := b.coils[id]
	return ev, ok
}

func (b *EvaporatorBank) Frostiest(threshold float64) (model.DuctID, float64, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var best model.DuctID
	var maxFrost float64
	found := false
	for id, ev := range b.coils {
		if ev.FrostIndex >= threshold && ev.FrostIndex > maxFrost {
			best = id
			maxFrost = ev.FrostIndex
			found = true
		}
	}
	return best, maxFrost, found
}

func (b *EvaporatorBank) RunDefrostCycle(ctx context.Context, id model.DuctID, meltRate float64) error {
	select {
	case <-ctx.Done():
		return model.Wrap("evaporator", "defrost", context.Cause(ctx))
	default:
	}
	b.mu.Lock()
	ev, ok := b.coils[id]
	b.mu.Unlock()
	if !ok {
		return model.Wrap("evaporator", "missing", model.ErrNotFound)
	}
	ev.EnableHeater(true)
	ev.DefrostProgress(meltRate)
	if ev.FrostIndex == 0 {
		ev.EnableHeater(false)
	}
	return nil
}

func (b *EvaporatorBank) Summary() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	parts := make([]string, 0, len(b.coils))
	for id, ev := range b.coils {
		parts = append(parts, fmt.Sprintf("%s:frost=%.2f heater=%v", id, ev.FrostIndex, ev.HeaterOn))
	}
	return fmt.Sprintf("evaporators=%d {%s}", len(b.coils), stringsJoin(parts, ","))
}

func stringsJoin(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += sep + parts[i]
	}
	return out
}
