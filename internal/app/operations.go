package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/reefctl/internal/model"
)

func (a *App) ValidateRefrigerantHigh(ctx context.Context, psi float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if psi <= 350.0 {
		return nil
	}
	// Raise a dedicated refrigerant alarm so operators see a refrigerant-class
	// detail rather than a generic unit fault, and can dispatch the high-side
	// handling branch. The subsequent unit-fault transition still raises
	// COMPRESSOR_TRIP, but the refrigerant event carries the high-pressure
	// semantic label and timestamp distinct from the trip receipt.
	if a.alarms != nil {
		_ = a.alarms.Raise(ctx, "REFRIGERANT_HIGH", model.UnitID(a.cfg.UnitID), 3)
	}
	return fmt.Errorf("refrigerant: %v", model.ErrRefrigerantHigh)
}

func (a *App) ConfirmZoneHold(ctx context.Context, anchor time.Time) error {
	if a.zoneWindow == nil {
		return model.Wrap("zone hold", "window", model.ErrZoneHold)
	}
	if err := a.zoneWindow.Require(anchor); err != nil {
		return fmt.Errorf("zone hold: %w", err)
	}
	_ = ctx
	return nil
}
