package model

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidID       = errors.New("reefctl: invalid identifier")
	ErrNotFound        = errors.New("reefctl: entity not found")
	ErrConflict        = errors.New("reefctl: state conflict")
	ErrInterlock       = errors.New("reefctl: interlock denied")
	ErrDefrostActive     = errors.New("reefctl: thermal hold active")
	ErrAirflowSetpoint    = errors.New("reefctl: flow setpoint violation")
	ErrCompressor      = errors.New("reefctl: compressor fault")
	ErrScheduleEmpty   = errors.New("reefctl: schedule empty")
	ErrContextCanceled = errors.New("reefctl: operation canceled")
)

type DomainError struct {
	Op   string
	Code string
	Err  error
}

func (e *DomainError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("reefctl %s [%s]: %v", e.Op, e.Code, e.Err)
	}
	return fmt.Sprintf("reefctl %s [%s]", e.Op, e.Code)
}

func (e *DomainError) Unwrap() error { return e.Err }

func Wrap(op, code string, err error) error {
	if err == nil {
		return nil
	}
	return &DomainError{Op: op, Code: code, Err: err}
}

func Is(err, target error) bool { return errors.Is(err, target) }
func As(err error, target any) bool { return errors.As(err, target) }