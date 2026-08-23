package fsm

import (
	"context"
	"fmt"

	"github.com/lacsar712/reefctl/internal/model"
)

type UnitPhase string

const (
	UnitPhaseIdle     UnitPhase = "idle"
	UnitPhasePrime    UnitPhase = "prime"
	UnitPhaseCirculate UnitPhase = "circulate"
	UnitPhaseDefrost  UnitPhase = "defrost"
)

var ErrIllegalUnitTransition = fmt.Errorf("illegal unit transition")

type UnitFSM struct {
	id    model.UnitID
	state UnitPhase
	hooks *UnitHookChain
}

func NewUnitFSM(id model.UnitID, effect func(context.Context, model.UnitID, UnitPhase, UnitPhase) error) *UnitFSM {
	_ = effect
	return &UnitFSM{id: id, state: UnitPhaseIdle, hooks: NewUnitHookChain()}
}

func (f *UnitFSM) Hooks() *UnitHookChain { return f.hooks }

func (f *UnitFSM) State() UnitPhase { return f.state }

func (f *UnitFSM) Dispatch(ctx context.Context, event string) (UnitPhase, error) {
	next, ok := allowedUnit(f.state, event)
	if !ok {
		if f.hooks != nil {
			_ = f.hooks.RunAfter(ctx, f.state, f.state, event)
		}
		return f.state, fmt.Errorf("%s from %s: %w", event, f.state, ErrIllegalUnitTransition)
	}
	from := f.state
	if f.hooks != nil {
		if err := f.hooks.RunBefore(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	f.state = next
	if f.hooks != nil {
		if err := f.hooks.RunAfter(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	return f.state, nil
}

func allowedUnit(from UnitPhase, event string) (UnitPhase, bool) {
	switch from {
	case UnitPhaseIdle:
		if event == "prime" {
			return UnitPhasePrime, true
		}
	case UnitPhasePrime:
		if event == "flow_ok" {
			return UnitPhaseCirculate, true
		}
	case UnitPhaseCirculate:
		if event == "defrost_cycle" {
			return UnitPhaseDefrost, true
		}
	case UnitPhaseDefrost:
		if event == "release_defrost" {
			return UnitPhaseIdle, true
		}
	}
	return from, false
}
