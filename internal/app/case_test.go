package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/config"
	"github.com/lacsar712/reefctl/internal/defrost"
	"github.com/lacsar712/reefctl/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	win := defrost.NewWindow(a.clk.Now(), time.Minute, 6.5)
	a.unit.ArmDefrostCycle(win)
	err = a.HandleDefrostPending(context.Background(), model.UnitID(a.cfg.UnitID))
	if err == nil {
		t.Fatal("expected defrost pending error")
	}
	if !errors.Is(err, model.ErrDefrostPending) {
		t.Fatalf("expected ErrDefrostPending, got %v", err)
	}
}
