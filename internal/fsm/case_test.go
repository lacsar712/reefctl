package fsm

import (
	"context"
	"testing"

	"github.com/lacsar712/reefctl/internal/model"
)

func TestCase(t *testing.T) {
	var pulses int
	UnitCompressorPulse = func() { pulses++ }
	f := NewUnitFSM(model.UnitID("reef-unit-7"), nil)
	RegisterUnitCompressorHook(f.Hooks())
	_, err := f.Dispatch(context.Background(), "flow_ok")
	if err == nil {
		t.Fatal("expected illegal unit transition error")
	}
	if pulses != 0 {
		t.Fatalf("illegal unit transition should not pulse compressor, got %d", pulses)
	}
}
