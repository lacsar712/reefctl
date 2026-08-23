package alarms

import (
	"sync"

	"github.com/lacsar712/reefctl/internal/model"
)

type Registry struct {
	mu     sync.RWMutex
	codes  map[model.AlarmCode]string
	raised map[model.AlarmCode]int
}

func NewRegistry() *Registry {
	return &Registry{
		codes: map[model.AlarmCode]string{
			"FLOW_LOW":        "airflow below setpoint",
			"DEFROST_OVERRUN": "defrost cycle window exceeded",
			"COMPRESSOR_TRIP": "compressor tripped",
			"VALVE_STUCK":     "damper interlock timeout",
		},
		raised: make(map[model.AlarmCode]int),
	}
}

func (r *Registry) Describe(code model.AlarmCode) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msg, ok := r.codes[code]
	return msg, ok
}

func (r *Registry) Register(code model.AlarmCode, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.codes[code] = message
}

func (r *Registry) Count(code model.AlarmCode) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.raised[code]
}

func (r *Registry) MarkRaised(code model.AlarmCode) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.raised[code]++
}
