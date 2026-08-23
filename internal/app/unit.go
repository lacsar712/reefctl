package app

import (
	"context"

	"github.com/lacsar712/reefctl/internal/model"
)

func (a *App) BeginUnitScope(ctx context.Context, unit model.UnitID) (context.Context, context.CancelFunc) {
	if unit == "" {
		unit = model.UnitID(a.cfg.UnitID)
	}
	a.unitMu.Lock()
	if cancel, ok := a.unitCancels[unit]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	a.unitCancels[unit] = cancel
	a.unitMu.Unlock()
	release := func() {
		a.unitMu.Lock()
		delete(a.unitCancels, unit)
		a.unitMu.Unlock()
		cancel()
	}
	return child, release
}

func (a *App) RunUnit(ctx context.Context, unit model.UnitID, fn func(context.Context) error) error {
	unitCtx, release := a.BeginUnitScope(ctx, unit)
	defer release()
	return fn(unitCtx)
}
