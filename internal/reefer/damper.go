package reefer

import (
	"context"
	"time"

	"github.com/lacsar712/reefctl/internal/interlock"
	"github.com/lacsar712/reefctl/internal/model"
)

type DamperActuator struct {
	lock *interlock.ValveLock
	pos  map[model.ValveID]model.ValvePosition
}

func NewDamperActuator(lock *interlock.ValveLock) *DamperActuator {
	return &DamperActuator{lock: lock, pos: make(map[model.ValveID]model.ValvePosition)}
}

func (v *DamperActuator) Position(id model.ValveID) model.ValvePosition {
	if p, ok := v.pos[id]; ok {
		return p
	}
	return model.ValveClosed
}

func (v *DamperActuator) Open(ctx context.Context, id model.ValveID, ttlSec int) error {
	return v.lock.WithLease(ctx, id, time.Duration(ttlSec)*time.Second, func() error {
		v.pos[id] = model.ValveOpen
		return nil
	})
}

func (v *DamperActuator) Close(ctx context.Context, id model.ValveID, ttlSec int) error {
	return v.lock.WithLease(ctx, id, time.Duration(ttlSec)*time.Second, func() error {
		v.pos[id] = model.ValveClosed
		return nil
	})
}
