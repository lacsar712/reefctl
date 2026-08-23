package fsm

import (
	"fmt"

	"github.com/lacsar712/reefctl/internal/model"
)

type Transition struct {
	From  model.UnitState
	To    model.UnitState
	Event string
}

var reeferTransitions = []Transition{
	{model.UnitIdle, model.UnitPriming, "prime"},
	{model.UnitPriming, model.UnitCirculate, "flow_ok"},
	{model.UnitCirculate, model.UnitDefrost, "defrost_cycle"},
	{model.UnitDefrost, model.UnitCirculate, "release_defrost"},
	{model.UnitCirculate, model.UnitIdle, "stop"},
	{model.UnitPriming, model.UnitFault, "fault"},
	{model.UnitCirculate, model.UnitFault, "fault"},
	{model.UnitDefrost, model.UnitFault, "fault"},
	{model.UnitFault, model.UnitShutdown, "shutdown"},
	{model.UnitIdle, model.UnitShutdown, "shutdown"},
}

func AllowedReefer(from model.UnitState, event string) (model.UnitState, bool) {
	for _, t := range reeferTransitions {
		if t.From == from && t.Event == event {
			return t.To, true
		}
	}
	return from, false
}

func MustReefer(from model.UnitState, event string) (model.UnitState, error) {
	to, ok := AllowedReefer(from, event)
	if !ok {
		return from, model.Wrap("reefer_fsm", "illegal_transition", fmt.Errorf("%s -> %s", from, event))
	}
	return to, nil
}

var compressorTransitions = []struct {
	from, to model.CompressorState
	event    string
}{
	{model.CompressorOff, model.CompressorStaging, "start"},
	{model.CompressorStaging, model.CompressorRun, "staged"},
	{model.CompressorRun, model.CompressorCoast, "stop"},
	{model.CompressorCoast, model.CompressorOff, "coast_done"},
	{model.CompressorRun, model.CompressorTrip, "trip"},
	{model.CompressorStaging, model.CompressorTrip, "trip"},
}

func AllowedCompressor(from model.CompressorState, event string) (model.CompressorState, bool) {
	for _, t := range compressorTransitions {
		if t.from == from && t.event == event {
			return t.to, true
		}
	}
	return from, false
}