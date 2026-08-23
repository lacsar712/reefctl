package fsm

import (
	"context"

	"github.com/lacsar712/reefctl/internal/model"
)

type ReeferSideEffect func(ctx context.Context, rack model.UnitID, from, to model.UnitState) error

type ReeferFSM struct {
	id       model.UnitID
	state    model.UnitState
	onChange ReeferSideEffect
}

func NewReeferFSM(id model.UnitID, effect ReeferSideEffect) *ReeferFSM {
	return &ReeferFSM{id: id, state: model.UnitIdle, onChange: effect}
}

func (f *ReeferFSM) State() model.UnitState { return f.state }

func (f *ReeferFSM) Apply(ctx context.Context, event string) error {
	next, err := MustReefer(f.state, event)
	if err != nil {
		return err
	}
	prev := f.state
	if f.onChange != nil {
		if err := f.onChange(ctx, f.id, prev, next); err != nil {
			return model.Wrap("reefer_fsm", "side_effect", err)
		}
	}
	f.state = next
	return nil
}