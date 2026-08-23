package model

import (
	"fmt"
	"strings"
)

type UnitID string
type ZoneID string
type DuctID string
type CompressorID string
type ValveID string
type SensorID string
type ScheduleID string
type AlarmCode string

func (id UnitID) String() string       { return string(id) }
func (id ZoneID) String() string       { return string(id) }
func (id DuctID) String() string   { return string(id) }
func (id CompressorID) String() string { return string(id) }
func (id ValveID) String() string      { return string(id) }
func (id SensorID) String() string     { return string(id) }
func (id ScheduleID) String() string   { return string(id) }
func (id AlarmCode) String() string    { return string(id) }

func ParseUnitID(raw string) (UnitID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return UnitID(raw), nil
}

func ParseZoneID(rack UnitID, index int) (ZoneID, error) {
	if rack == "" || index < 0 {
		return "", ErrInvalidID
	}
	return ZoneID(fmt.Sprintf("%s-slot-%02d", rack, index)), nil
}

func ParseDuctID(raw string) (DuctID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return DuctID(raw), nil
}

func ParseCompressorID(raw string) (CompressorID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return CompressorID(raw), nil
}

func ParseValveID(raw string) (ValveID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return ValveID(raw), nil
}

func ParseSensorID(raw string) (SensorID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return SensorID(raw), nil
}

func ParseScheduleID(raw string) (ScheduleID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return ScheduleID(raw), nil
}