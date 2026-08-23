package store

import (
	"sort"
	"time"

	"github.com/lacsar712/reefctl/internal/model"
)

type ScheduleStore struct{ mem *Memory }

func NewScheduleStore(mem *Memory) *ScheduleStore { return &ScheduleStore{mem: mem} }

func (ss *ScheduleStore) Save(s model.DefrostSchedule) {
	s.Version++
	ss.mem.PutSchedule(s.Clone())
}

func (ss *ScheduleStore) ActiveEntry(s model.DefrostSchedule, at time.Time) (model.DefrostScheduleEntry, bool) {
	clone := s.Clone()
	if len(clone.Entries) == 0 {
		return model.DefrostScheduleEntry{}, false
	}
	entries := append([]model.DefrostScheduleEntry(nil), clone.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Start.Before(entries[j].Start) })
	for _, e := range entries {
		if !at.Before(e.Start) && at.Before(e.End) {
			return e, true
		}
	}
	return model.DefrostScheduleEntry{}, false
}

func (ss *ScheduleStore) SnapshotClone(id model.ScheduleID) (model.DefrostSchedule, error) {
	s, ok := ss.mem.GetSchedule(id)
	if !ok {
		return model.DefrostSchedule{}, model.Wrap("schedule", "not_found", model.ErrNotFound)
	}
	return s.Clone(), nil
}

func MergeSchedules(dst model.DefrostSchedule, extra []model.DefrostScheduleEntry) model.DefrostSchedule {
	out := dst.Clone()
	out.Entries = append(out.Entries, extra...)
	sort.Slice(out.Entries, func(i, j int) bool { return out.Entries[i].Start.Before(out.Entries[j].Start) })
	return out
}