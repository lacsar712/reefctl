package reefer

import (
	"sync"

	"github.com/lacsar712/reefctl/internal/model"
)

type Duct struct {
	ID         model.DuctID
	Capacity   float64
	Primed     bool
	AirflowCFM float64
}

func (m *Duct) Ready() bool              { return m.Primed }
func (m *Duct) Prime()                 { m.Primed = true }
func (m *Duct) SetFlow(cfm float64)    { m.AirflowCFM = cfm }

type DuctRegistry struct {
	mu   sync.RWMutex
	data map[model.DuctID]*Duct
}

func NewDuctRegistry() *DuctRegistry {
	return &DuctRegistry{data: make(map[model.DuctID]*Duct)}
}

func (r *DuctRegistry) Add(m *Duct) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[m.ID] = m
}

func (r *DuctRegistry) Get(id model.DuctID) (*Duct, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.data[id]
	return m, ok
}

func (r *DuctRegistry) All() []*Duct {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Duct, 0, len(r.data))
	for _, m := range r.data {
		out = append(out, m)
	}
	return out
}
