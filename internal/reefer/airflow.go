package reefer

import (
	"context"

	"github.com/lacsar712/reefctl/internal/model"
)

type AirflowController struct {
	setpoint model.AirflowSetpoint
	actual   float64
}

func NewAirflowController(sp model.AirflowSetpoint) *AirflowController {
	return &AirflowController{setpoint: sp}
}

func (f *AirflowController) SetSetpoint(sp model.AirflowSetpoint) { f.setpoint = sp }
func (f *AirflowController) Observe(cfm float64)                  { f.actual = cfm }
func (f *AirflowController) Actual() float64                      { return f.actual }
func (f *AirflowController) Setpoint() model.AirflowSetpoint      { return f.setpoint }

func (f *AirflowController) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return model.Wrap("airflow", "canceled", context.Cause(ctx))
	default:
	}
	if !f.setpoint.Within(f.actual) {
		return model.Wrap("airflow", "setpoint", model.ErrAirflowSetpoint)
	}
	return nil
}
