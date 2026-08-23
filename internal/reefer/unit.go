package reefer

import (
	"context"
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/compressor"
	"github.com/lacsar712/reefctl/internal/config"
	"github.com/lacsar712/reefctl/internal/defrost"
	"github.com/lacsar712/reefctl/internal/model"
	"github.com/lacsar712/reefctl/internal/store"
)

type ReeferUnit struct {
	cfg         config.Config
	clk         clock.Clock
	coordinator *compressor.Coordinator
	ducts       *DuctRegistry
	flow        map[model.DuctID]*AirflowController
	cycle       *defrost.CycleController
	sensors     *defrost.SensorBank
	store       *store.Memory
}

func NewReeferUnit(cfg config.Config, clk clock.Clock, mem *store.Memory) *ReeferUnit {
	return &ReeferUnit{
		cfg:         cfg,
		clk:         clk,
		coordinator: compressor.NewCoordinator(cfg, clk),
		ducts:       NewDuctRegistry(),
		flow:        make(map[model.DuctID]*AirflowController),
		cycle:       defrost.NewCycleController(clk),
		sensors:     defrost.NewSensorBank(clk),
		store:       mem,
	}
}

func (p *ReeferUnit) PrimeDuct(ctx context.Context, id model.DuctID) error {
	m, ok := p.ducts.Get(id)
	if !ok {
		return model.Wrap("reefer", "duct", model.ErrNotFound)
	}
	deadline := p.clk.Now().Add(time.Duration(p.cfg.DuctPrimeSec) * time.Second)
	primed := false
	for p.clk.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return model.Wrap("reefer", "prime", context.Cause(ctx))
		default:
		}
		if !primed {
			m.Prime()
			primed = true
		}
		if m.Ready() {
			return nil
		}
		if pc, ok := p.clk.(*clock.ProcessClock); ok {
			pc.Step()
		} else {
			time.Sleep(5 * time.Millisecond)
		}
	}
	return model.Wrap("reefer", "prime_timeout", model.ErrConflict)
}

func (p *ReeferUnit) BindFlow(id model.DuctID, sp model.AirflowSetpoint) {
	p.flow[id] = NewAirflowController(sp)
}

func (p *ReeferUnit) ObserveFlow(id model.DuctID, cfm float64) error {
	fc, ok := p.flow[id]
	if !ok {
		return model.Wrap("reefer", "flow_bind", model.ErrNotFound)
	}
	fc.Observe(cfm)
	return nil
}

func (p *ReeferUnit) ValidateFlows(ctx context.Context) error {
	for id, fc := range p.flow {
		if err := fc.Validate(ctx); err != nil {
			return model.Wrap("reefer", string(id), err)
		}
	}
	return nil
}

func (p *ReeferUnit) ArmDefrostCycle(w defrost.Window) { p.cycle.Arm(w) }
func (p *ReeferUnit) DefrostActive() bool               { return p.cycle.Active() }
func (p *ReeferUnit) Coordinator() *compressor.Coordinator { return p.coordinator }
func (p *ReeferUnit) Ducts() *DuctRegistry              { return p.ducts }
func (p *ReeferUnit) Sensors() *defrost.SensorBank      { return p.sensors }
