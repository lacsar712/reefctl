package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/reefctl/internal/model"
)

var CalibrateProbe func(ctx context.Context) error

func (a *App) CalibrateZone(ctx context.Context, unit model.UnitID, holder string) error {
	if err := a.unitLeases.Require(unit, holder, 30*time.Second); err != nil {
		return err
	}
	if CalibrateProbe != nil {
		if err := CalibrateProbe(ctx); err != nil {
			a.unitLeases.ReleaseHolder(unit, holder)
			return fmt.Errorf("calibrate: %w", err)
		}
	}
	a.unitLeases.ReleaseHolder(unit, holder)
	return nil
}
