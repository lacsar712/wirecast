package model

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidID       = errors.New("wirecast: invalid identifier")
	ErrNotFound        = errors.New("wirecast: entity not found")
	ErrConflict        = errors.New("wirecast: state conflict")
	ErrInterlock       = errors.New("wirecast: interlock denied")
	ErrMoistureHold    = errors.New("wirecast: moisture hold active")
	ErrAirflowSetpoint = errors.New("wirecast: airflow setpoint violation")
	ErrFanFault        = errors.New("wirecast: fan fault")
	ErrScheduleEmpty   = errors.New("wirecast: schedule empty")
	ErrGradient        = errors.New("wirecast: moisture gradient violation")
	ErrShellDrift   = errors.New("wirecast: moisture drift exceeded")
	ErrHotSpot    = errors.New("wirecast: heat overtemperature")
	ErrSolidHold    = errors.New("wirecast: gradient hold not satisfied")
	ErrContextCanceled = errors.New("wirecast: operation canceled")
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
		return fmt.Sprintf("wirecast %s [%s]: %v", e.Op, e.Code, e.Err)
	}
	return fmt.Sprintf("wirecast %s [%s]", e.Op, e.Code)
}

func (e *DomainError) Unwrap() error { return e.Err }

func Wrap(op, code string, err error) error {
	if err == nil {
		return nil
	}
	return &DomainError{Op: op, Code: code, Err: err}
}

func Is(err, target error) bool   { return errors.Is(err, target) }
func As(err error, target any) bool { return errors.As(err, target) }
