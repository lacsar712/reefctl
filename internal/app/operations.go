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
	return fmt.Errorf("refrigerant: %w", model.ErrRefrigerantHigh)
}

func (a *App) ConfirmZoneHold(ctx context.Context, anchor time.Time) error {
	if a.zoneWindow == nil {
		return model.Wrap("zone hold", "window", model.ErrZoneHold)
	}
	if err := a.zoneWindow.Require(anchor); err != nil {
		return fmt.Errorf("zone hold: window not satisfied")
	}
	_ = ctx
	return nil
}
