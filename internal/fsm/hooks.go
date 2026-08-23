package fsm

import (
	"context"
)

type UnitTransitionHook func(ctx context.Context, from, to UnitPhase, event string) error

type UnitHookChain struct {
	before []UnitTransitionHook
	after  []UnitTransitionHook
}

func NewUnitHookChain() *UnitHookChain { return &UnitHookChain{} }

func (h *UnitHookChain) OnBefore(fn UnitTransitionHook) { h.before = append(h.before, fn) }
func (h *UnitHookChain) OnAfter(fn UnitTransitionHook)  { h.after = append(h.after, fn) }

func (h *UnitHookChain) RunBefore(ctx context.Context, from, to UnitPhase, event string) error {
	for _, fn := range h.before {
		if err := fn(ctx, from, to, event); err != nil {
			return err
		}
	}
	return nil
}

func (h *UnitHookChain) RunAfter(ctx context.Context, from, to UnitPhase, event string) error {
	for _, fn := range h.after {
		if err := fn(ctx, from, to, event); err != nil {
			return err
		}
	}
	return nil
}

var UnitCompressorPulse func()

func RegisterUnitCompressorHook(chain *UnitHookChain) {
	chain.OnAfter(func(ctx context.Context, from, to UnitPhase, event string) error {
		_ = ctx
		_ = from
		_ = to
		_ = event
		if UnitCompressorPulse != nil {
			UnitCompressorPulse()
		}
		return nil
	})
}
