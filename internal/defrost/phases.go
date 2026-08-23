package defrost

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/model"
)

type Phase string

const (
	PhaseIdle    Phase = "idle"
	PhasePreheat Phase = "preheat"
	PhaseMelt    Phase = "melt"
	PhaseDrip    Phase = "drip"
	PhaseSettle  Phase = "settle"
)

type PhaseController struct {
	mu             sync.Mutex
	clk            clock.Clock
	phase          Phase
	window         Window
	heater         *HeaterController
	duct           model.DuctID
	frostThreshold float64
	meltRate       float64
	dripTarget     float64
}

func NewPhaseController(clk clock.Clock, heater *HeaterController, duct model.DuctID) *PhaseController {
	return &PhaseController{
		clk:            clk,
		heater:         heater,
		duct:           duct,
		phase:          PhaseIdle,
		frostThreshold: 2.0,
		meltRate:       0.5,
		dripTarget:     4.0,
	}
}

func (p *PhaseController) Begin(w Window) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.window = w
	p.phase = PhasePreheat
}

func (p *PhaseController) Phase() Phase {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.phase
}

func (p *PhaseController) Active() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.phase != PhaseIdle && p.window.Active(p.clk)
}

func (p *PhaseController) Advance(ctx context.Context, frostIndex float64, reading model.TempReading) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.phase == PhaseIdle {
		return nil
	}
	select {
	case <-ctx.Done():
		return model.Wrap("defrost_phase", "canceled", context.Cause(ctx))
	default:
	}
	switch p.phase {
	case PhasePreheat:
		if err := p.heater.Engage(p.duct); err != nil {
			return err
		}
		if reading.Celsius >= p.window.TargetC-1 {
			p.phase = PhaseMelt
		}
	case PhaseMelt:
		if frostIndex <= 0 {
			p.heater.Release(p.duct)
			p.phase = PhaseDrip
		}
	case PhaseDrip:
		if reading.Celsius <= p.dripTarget {
			p.phase = PhaseSettle
		}
	case PhaseSettle:
		if p.window.WithinTarget(reading) || !p.window.Active(p.clk) {
			p.phase = PhaseIdle
			p.heater.Release(p.duct)
		}
	}
	return nil
}

func (p *PhaseController) ForceRelease() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.phase = PhaseIdle
	p.heater.Release(p.duct)
}

func (p *PhaseController) Summary() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return fmt.Sprintf("defrost_phase duct=%s phase=%s target=%.1fC", p.duct, p.phase, p.window.TargetC)
}
