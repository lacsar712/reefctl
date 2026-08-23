package compressor

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/config"
	"github.com/lacsar712/reefctl/internal/fsm"
	"github.com/lacsar712/reefctl/internal/model"
)

type Coordinator struct {
	mu    sync.Mutex
	cfg   config.Config
	clk   clock.Clock
	units map[model.CompressorID]*fsm.CompressorFSM
	log   []string
}

func NewCoordinator(cfg config.Config, clk clock.Clock) *Coordinator {
	return &Coordinator{cfg: cfg, clk: clk, units: make(map[model.CompressorID]*fsm.CompressorFSM)}
}

func (c *Coordinator) Register(id model.CompressorID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.units[id]; ok {
		return
	}
	effect := func(ctx context.Context, cid model.CompressorID, from, to model.CompressorState) error {
		c.log = append(c.log, fmt.Sprintf("%s %s->%s", cid, from, to))
		if to == model.CompressorTrip {
			return model.Wrap("compressor", "trip", model.ErrCompressor)
		}
		return nil
	}
	c.units[id] = fsm.NewCompressorFSM(c.clk, id, effect)
}

func (c *Coordinator) Start(ctx context.Context, id model.CompressorID) error {
	c.mu.Lock()
	unit, ok := c.units[id]
	c.mu.Unlock()
	if !ok {
		return model.Wrap("compressor", "missing", model.ErrNotFound)
	}
	if err := unit.Apply(ctx, "start"); err != nil {
		return err
	}
	return unit.Apply(ctx, "staged")
}

func (c *Coordinator) Stop(ctx context.Context, id model.CompressorID) error {
	c.mu.Lock()
	unit, ok := c.units[id]
	c.mu.Unlock()
	if !ok {
		return model.Wrap("compressor", "missing", model.ErrNotFound)
	}
	if !unit.CanStop(c.cfg.CompressorMinRun) {
		return model.Wrap("compressor", "min_run", model.ErrConflict)
	}
	if err := unit.Apply(ctx, "stop"); err != nil {
		return err
	}
	return unit.Apply(ctx, "coast_done")
}

func (c *Coordinator) States() map[model.CompressorID]model.CompressorState {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[model.CompressorID]model.CompressorState, len(c.units))
	for id, u := range c.units {
		out[id] = u.State()
	}
	return out
}