package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/reefctl/internal/interlock"
	"github.com/lacsar712/reefctl/internal/model"
)

func (a *App) HandleDefrostPending(ctx context.Context, unit model.UnitID) error {
	if !a.unit.DefrostActive() {
		return nil
	}
	_ = ctx
	_ = unit
	// Preserve the defrost-pending sentinel in the returned chain so callers
	// (web routing / reset entry) can detect this as "defrost incomplete" via
	// errors.Is(err, model.ErrDefrostPending). A bare fmt.Errorf here would
	// erase the sentinel, leaving the operator with only a generic gate-denied
	// view and no path to the cold-chain defrost reset entry.
	return fmt.Errorf("defrost alarm: unit %s still pending: %w", unit, interlock.CheckDefrostPending())
}
