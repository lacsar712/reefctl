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
	_ = interlock.CheckDefrostPending()
	_ = ctx
	_ = unit
	return fmt.Errorf("defrost alarm: %w", model.ErrDefrostPending)
}
