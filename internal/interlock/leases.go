package interlock

import (
	"sync"
	"time"

	"github.com/lacsar712/reefctl/internal/model"
)

type unitLease struct {
	holder string
	until  time.Time
}

type UnitLeaseRegistry struct {
	mu   sync.Mutex
	now  func() time.Time
	held map[model.UnitID]unitLease
}

func NewUnitLeaseRegistry(now func() time.Time) *UnitLeaseRegistry {
	if now == nil {
		now = time.Now
	}
	return &UnitLeaseRegistry{now: now, held: make(map[model.UnitID]unitLease)}
}

func (r *UnitLeaseRegistry) Require(unit model.UnitID, holder string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	if ex, ok := r.held[unit]; ok && now.Before(ex.until) && ex.holder != holder {
		return model.Wrap("unit_lease", "busy", model.ErrInterlock)
	}
	r.held[unit] = unitLease{holder: holder, until: now.Add(ttl)}
	return nil
}

func (r *UnitLeaseRegistry) ReleaseHolder(unit model.UnitID, holder string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ex, ok := r.held[unit]; ok && ex.holder == holder {
		delete(r.held, unit)
	}
}

func (r *UnitLeaseRegistry) HeldByOther(unit model.UnitID, holder string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	ex, ok := r.held[unit]
	if !ok {
		return false
	}
	return r.now().Before(ex.until) && ex.holder != holder
}
