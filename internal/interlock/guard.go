package interlock

import (
	"github.com/lacsar712/reefctl/internal/model"
)

type Guard struct {
	allowed map[model.ZoneID]model.DuctID
}

func NewGuard(pairs map[model.ZoneID]model.DuctID) *Guard {
	cp := make(map[model.ZoneID]model.DuctID, len(pairs))
	for k, v := range pairs {
		cp[k] = v
	}
	return &Guard{allowed: cp}
}

func (g *Guard) Permit(slot model.ZoneID, manifold model.DuctID) error {
	want, ok := g.allowed[slot]
	if !ok {
		return model.Wrap("interlock", "unknown_slot", model.ErrNotFound)
	}
	if want != manifold {
		return model.Wrap("interlock", "manifold_mismatch", model.ErrInterlock)
	}
	return nil
}

func (g *Guard) SlotsFor(manifold model.DuctID) []model.ZoneID {
	var out []model.ZoneID
	for slot, m := range g.allowed {
		if m == manifold {
			out = append(out, slot)
		}
	}
	return out
}