package reefer

import (
	"fmt"
	"sort"

	"github.com/lacsar712/reefctl/internal/model"
)

type Partition struct {
	ID       string
	Slots    []model.ZoneID
	Duct     model.DuctID
	SharePct float64
}

type PartitionTable struct {
	unit       model.UnitID
	partitions []Partition
}

func NewPartitionTable(unit model.UnitID) *PartitionTable {
	return &PartitionTable{unit: unit}
}

func (t *PartitionTable) BuildFromZones(zones *ZoneTable, guard *CargoMonitor) error {
	slots := zones.Slots()
	if len(slots) == 0 {
		return model.Wrap("partition", "empty", model.ErrNotFound)
	}
	byDuct := make(map[model.DuctID][]model.ZoneID)
	for _, s := range slots {
		if !s.Enabled {
			continue
		}
		byDuct[s.Duct] = append(byDuct[s.Duct], s.ID)
	}
	if len(byDuct) == 0 {
		return model.Wrap("partition", "no_enabled", model.ErrConflict)
	}
	t.partitions = t.partitions[:0]
	idx := 0
	for duct, zoneIDs := range byDuct {
		sort.Slice(zoneIDs, func(i, j int) bool { return zoneIDs[i] < zoneIDs[j] })
		share := 100.0 / float64(len(byDuct))
		if guard != nil {
			share = t.weightedShare(zoneIDs, guard, byDuct)
		}
		t.partitions = append(t.partitions, Partition{
			ID:       fmt.Sprintf("part-%d", idx),
			Slots:    zoneIDs,
			Duct:     duct,
			SharePct: share,
		})
		idx++
	}
	sort.Slice(t.partitions, func(i, j int) bool { return t.partitions[i].ID < t.partitions[j].ID })
	return t.normalizeShares()
}

func (t *PartitionTable) weightedShare(zoneIDs []model.ZoneID, guard *CargoMonitor, all map[model.DuctID][]model.ZoneID) float64 {
	violations := make(map[model.ZoneID]struct{}, len(guard.Violations()))
	for _, id := range guard.Violations() {
		violations[id] = struct{}{}
	}
	var weight float64
	for _, id := range zoneIDs {
		weight++
		if _, hot := violations[id]; hot {
			weight += 2
		}
	}
	total := 0.0
	for _, ids := range all {
		total += float64(len(ids))
	}
	if total == 0 {
		return 0
	}
	return weight / total * 100
}

func (t *PartitionTable) normalizeShares() error {
	if len(t.partitions) == 0 {
		return model.Wrap("partition", "empty", model.ErrNotFound)
	}
	var sum float64
	for i := range t.partitions {
		sum += t.partitions[i].SharePct
	}
	if sum <= 0 {
		each := 100.0 / float64(len(t.partitions))
		for i := range t.partitions {
			t.partitions[i].SharePct = each
		}
		return nil
	}
	for i := range t.partitions {
		t.partitions[i].SharePct = t.partitions[i].SharePct / sum * 100
	}
	return nil
}

func (t *PartitionTable) Partitions() []Partition {
	out := make([]Partition, len(t.partitions))
	copy(out, t.partitions)
	return out
}

func (t *PartitionTable) DuctForZone(zone model.ZoneID) (model.DuctID, bool) {
	for _, p := range t.partitions {
		for _, id := range p.Slots {
			if id == zone {
				return p.Duct, true
			}
		}
	}
	return "", false
}

func (t *PartitionTable) ShareForDuct(duct model.DuctID) float64 {
	for _, p := range t.partitions {
		if p.Duct == duct {
			return p.SharePct
		}
	}
	return 0
}
