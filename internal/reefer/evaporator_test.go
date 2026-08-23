package reefer

import (
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
)

func TestEvaporatorFrostCycle(t *testing.T) {
	clk := clock.NewProcessClock(clkBase(), time.Millisecond)
	bank := NewEvaporatorBank(clk)
	ev := bank.Register("duct-evap")
	ev.AccumulateFrost(2.5)
	if !ev.NeedsDefrost(2.0) {
		t.Fatal("should need defrost")
	}
	if err := bank.RunDefrostCycle(t.Context(), "duct-evap", 3.0); err != nil {
		t.Fatal(err)
	}
	if ev.FrostIndex != 0 {
		t.Fatalf("frost=%f", ev.FrostIndex)
	}
}

func TestEvaporatorBankFrostiest(t *testing.T) {
	clk := clock.NewProcessClock(clkBase(), time.Millisecond)
	bank := NewEvaporatorBank(clk)
	bank.Register("a").AccumulateFrost(1.0)
	bank.Register("b").AccumulateFrost(3.0)
	id, frost, ok := bank.Frostiest(2.0)
	if !ok || id != "b" || frost != 3.0 {
		t.Fatalf("got %s %.1f %v", id, frost, ok)
	}
}

func clkBase() time.Time {
	return time.Unix(0, 0)
}
