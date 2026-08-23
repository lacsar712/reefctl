package reefer

import (
	"sort"

	"github.com/lacsar712/reefctl/internal/model"
)

type Router struct{ routes []model.DuctRoute }

func NewRouter(routes []model.DuctRoute) *Router {
	cp := append([]model.DuctRoute(nil), routes...)
	sort.Slice(cp, func(i, j int) bool {
		if cp[i].Priority == cp[j].Priority {
			return cp[i].From < cp[j].From
		}
		return cp[i].Priority > cp[j].Priority
	})
	return &Router{routes: cp}
}

func (r *Router) Path(from, to model.DuctID) ([]model.DuctRoute, bool) {
	for _, route := range r.routes {
		if route.From == from && route.To == to {
			return []model.DuctRoute{route}, true
		}
	}
	return nil, false
}

func (r *Router) ValvesFor(path []model.DuctRoute) []model.ValveID {
	out := make([]model.ValveID, 0, len(path))
	for _, p := range path {
		out = append(out, p.Valve)
	}
	return out
}

func (r *Router) Reachable(from model.DuctID) []model.DuctID {
	seen := make(map[model.DuctID]struct{})
	for _, route := range r.routes {
		if route.From == from {
			seen[route.To] = struct{}{}
		}
	}
	out := make([]model.DuctID, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}