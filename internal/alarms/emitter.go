package alarms

import (
	"context"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/model"
)

type Emitter struct {
	reg      *Registry
	clk      clock.Clock
	buffer   chan model.AlarmEvent
	capacity int
}

func NewEmitter(reg *Registry, clk clock.Clock, capacity int) *Emitter {
	if capacity <= 0 {
		capacity = 32
	}
	return &Emitter{reg: reg, clk: clk, buffer: make(chan model.AlarmEvent, capacity), capacity: capacity}
}

func (e *Emitter) Raise(ctx context.Context, code model.AlarmCode, unit model.UnitID, severity int) error {
	select {
	case <-ctx.Done():
		return model.Wrap("alarms", "canceled", context.Cause(ctx))
	default:
	}
	msg, ok := e.reg.Describe(code)
	if !ok {
		msg = "unknown alarm"
	}
	ev := model.AlarmEvent{Code: code, Message: msg, Unit: unit, RaisedAt: e.clk.Now(), Severity: severity}
	e.reg.MarkRaised(code)
	select {
	case e.buffer <- ev:
		return nil
	default:
		return model.Wrap("alarms", "buffer_full", model.ErrConflict)
	}
}

func (e *Emitter) Drain(max int) []model.AlarmEvent {
	if max <= 0 {
		max = e.capacity
	}
	var out []model.AlarmEvent
	for i := 0; i < max; i++ {
		select {
		case ev := <-e.buffer:
			out = append(out, ev)
		default:
			return out
		}
	}
	return out
}
