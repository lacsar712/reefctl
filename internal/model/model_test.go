package model

import (
	"errors"
	"testing"
)

func TestParseUnitID(t *testing.T) {
	id, err := ParseUnitID("rack-a")
	if err != nil || id != "rack-a" {
		t.Fatalf("parse rack: %v %q", err, id)
	}
	_, err = ParseUnitID("  ")
	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected invalid id")
	}
}

func TestAirflowSetpointWithin(t *testing.T) {
	sp := AirflowSetpoint{CubicFeetPerMinute: 10, TolerancePct: 10}
	if !sp.Within(10.5) {
		t.Fatal("within tolerance")
	}
	if sp.Within(12) {
		t.Fatal("outside tolerance")
	}
}

func TestScheduleClone(t *testing.T) {
	s := DefrostSchedule{ID: "s1", Entries: []DefrostScheduleEntry{{ID: "e1"}}}
	c := s.Clone()
	c.Entries[0].ID = "mutated"
	if s.Entries[0].ID == "mutated" {
		t.Fatal("clone should be deep for entries slice")
	}
}

func TestWrapUnwrap(t *testing.T) {
	err := Wrap("op", "code", ErrNotFound)
	var de *DomainError
	if !errors.As(err, &de) {
		t.Fatal("expected domain error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatal("unwrap chain")
	}
}