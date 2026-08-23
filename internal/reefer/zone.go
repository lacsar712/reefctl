package reefer

import (
	"fmt"

	"github.com/lacsar712/reefctl/internal/model"
)

type Slot struct {
	ID       model.ZoneID
	Duct model.DuctID
	Enabled  bool
	Index    int
}

func NewZone(rack model.UnitID, index int, manifold model.DuctID) (Slot, error) {
	id, err := model.ParseZoneID(rack, index)
	if err != nil {
		return Slot{}, err
	}
	return Slot{ID: id, Duct: manifold, Index: index}, nil
}

type ZoneTable struct {
	rack  model.UnitID
	slots []Slot
}

func NewZoneTable(rack model.UnitID, count int, defaultDuct model.DuctID) (*ZoneTable, error) {
	if count <= 0 {
		return nil, fmt.Errorf("slot count")
	}
	t := &ZoneTable{rack: rack}
	for i := 0; i < count; i++ {
		s, err := NewZone(rack, i, defaultDuct)
		if err != nil {
			return nil, err
		}
		t.slots = append(t.slots, s)
	}
	return t, nil
}

func (t *ZoneTable) Slots() []Slot {
	out := make([]Slot, len(t.slots))
	copy(out, t.slots)
	return out
}

func (t *ZoneTable) Assign(index int, manifold model.DuctID) error {
	if index < 0 || index >= len(t.slots) {
		return model.ErrNotFound
	}
	t.slots[index].Duct = manifold
	return nil
}

func (t *ZoneTable) Enable(index int, on bool) error {
	if index < 0 || index >= len(t.slots) {
		return model.ErrNotFound
	}
	t.slots[index].Enabled = on
	return nil
}

func (t *ZoneTable) EnabledCount() int {
	n := 0
	for _, s := range t.slots {
		if s.Enabled {
			n++
		}
	}
	return n
}