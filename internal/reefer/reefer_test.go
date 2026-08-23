package reefer

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/config"
	"github.com/lacsar712/reefctl/internal/interlock"
	"github.com/lacsar712/reefctl/internal/model"
	"github.com/lacsar712/reefctl/internal/store"
)

func TestReeferPrimeAndFlow(t *testing.T) {
	cfg := config.Default()
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	plant := NewReeferUnit(cfg, clk, store.NewMemory())
	plant.Ducts().Add(&Duct{ID: "duct1"})
	plant.BindFlow("duct1", model.AirflowSetpoint{CubicFeetPerMinute: 10, TolerancePct: 10})
	if err := plant.PrimeDuct(context.Background(), "duct1"); err != nil {
		t.Fatal(err)
	}
	_ = plant.ObserveFlow("duct1", 10)
	if err := plant.ValidateFlows(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestDamperActuator(t *testing.T) {
	v := NewDamperActuator(interlock.NewValveLock(time.Now))
	if err := v.Open(context.Background(), "v1", 1); err != nil {
		t.Fatal(err)
	}
	if v.Position("v1") != model.ValveOpen {
		t.Fatal("open")
	}
}
