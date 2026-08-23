package store

import (
	"time"

	"github.com/lacsar712/reefctl/internal/model"
)

type SnapshotBuilder struct {
	id    model.UnitID
	state model.UnitState
	slots []model.ZoneAssignment
	comp  []model.CompressorID
}

func NewSnapshotBuilder(id model.UnitID) *SnapshotBuilder {
	return &SnapshotBuilder{id: id, state: model.UnitIdle}
}

func (b *SnapshotBuilder) State(s model.UnitState) *SnapshotBuilder {
	b.state = s
	return b
}

func (b *SnapshotBuilder) Slot(a model.ZoneAssignment) *SnapshotBuilder {
	b.slots = append(b.slots, a)
	return b
}

func (b *SnapshotBuilder) Compressor(id model.CompressorID) *SnapshotBuilder {
	b.comp = append(b.comp, id)
	return b
}

func (b *SnapshotBuilder) Build(at time.Time) model.UnitSnapshot {
	slots := make([]model.ZoneAssignment, len(b.slots))
	copy(slots, b.slots)
	comp := make([]model.CompressorID, len(b.comp))
	copy(comp, b.comp)
	return model.UnitSnapshot{ID: b.id, State: b.state, Slots: slots, Compressors: comp, UpdatedAt: at}
}

func DiffSlots(before, after model.UnitSnapshot) []model.ZoneID {
	index := make(map[model.ZoneID]model.ZoneAssignment)
	for _, s := range before.Slots {
		index[s.Slot] = s
	}
	var changed []model.ZoneID
	for _, s := range after.Slots {
		prev, ok := index[s.Slot]
		if !ok || prev.LastFlow != s.LastFlow || prev.Enabled != s.Enabled {
			changed = append(changed, s.Slot)
		}
	}
	return changed
}