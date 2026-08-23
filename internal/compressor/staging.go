package compressor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/reefctl/internal/model"
)

type DefrostGate func() bool

type StagingInterlock struct {
	mu            sync.Mutex
	coordinator   *Coordinator
	defrostActive DefrostGate
	lastStage     map[model.CompressorID]time.Time
	minGap        time.Duration
	order         []model.CompressorID
	now           func() time.Time
	blocked       map[model.CompressorID]string
}

func NewStagingInterlock(co *Coordinator, gate DefrostGate, minGap time.Duration, order []model.CompressorID, now func() time.Time) *StagingInterlock {
	if now == nil {
		now = time.Now
	}
	cp := append([]model.CompressorID(nil), order...)
	return &StagingInterlock{
		coordinator:   co,
		defrostActive: gate,
		lastStage:     make(map[model.CompressorID]time.Time),
		minGap:        minGap,
		order:         cp,
		now:           now,
		blocked:       make(map[model.CompressorID]string),
	}
}

func (s *StagingInterlock) checkDefrost() error {
	if s.defrostActive != nil && s.defrostActive() {
		return model.Wrap("staging", "defrost_active", model.ErrDefrostActive)
	}
	return nil
}

func (s *StagingInterlock) checkGap(id model.CompressorID) error {
	if s.minGap <= 0 {
		return nil
	}
	for _, prev := range s.order {
		if prev == id {
			break
		}
		if last, ok := s.lastStage[prev]; ok {
			if s.now().Sub(last) < s.minGap {
				return model.Wrap("staging", "gap", model.ErrInterlock)
			}
		}
		if states := s.coordinator.States(); states[prev] == model.CompressorStaging {
			return model.Wrap("staging", "peer_staging", model.ErrInterlock)
		}
	}
	return nil
}

func (s *StagingInterlock) checkRunningPeers(id model.CompressorID) error {
	states := s.coordinator.States()
	for _, peer := range s.order {
		if peer == id {
			continue
		}
		st := states[peer]
		if st == model.CompressorRun || st == model.CompressorStaging {
			return model.Wrap("staging", "peer_running", model.ErrInterlock)
		}
	}
	return nil
}

func (s *StagingInterlock) RequestStart(ctx context.Context, id model.CompressorID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if reason, blocked := s.blocked[id]; blocked {
		return model.Wrap("staging", reason, model.ErrInterlock)
	}
	if err := s.checkDefrost(); err != nil {
		return err
	}
	if err := s.checkGap(id); err != nil {
		return err
	}
	if err := s.checkRunningPeers(id); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return model.Wrap("staging", "canceled", context.Cause(ctx))
	default:
	}
	if err := s.coordinator.Start(ctx, id); err != nil {
		return err
	}
	s.lastStage[id] = s.now()
	return nil
}

func (s *StagingInterlock) Block(id model.CompressorID, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blocked[id] = reason
}

func (s *StagingInterlock) Unblock(id model.CompressorID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.blocked, id)
}

func (s *StagingInterlock) StageSequence(ctx context.Context) error {
	for _, id := range s.order {
		if err := s.RequestStart(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *StagingInterlock) TripAll(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range s.order {
		s.blocked[id] = "trip"
		_ = s.coordinator.Stop(ctx, id)
	}
}

func (s *StagingInterlock) Status() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	states := s.coordinator.States()
	parts := make([]string, 0, len(s.order))
	for _, id := range s.order {
		st := states[id]
		if reason, ok := s.blocked[id]; ok {
			parts = append(parts, fmt.Sprintf("%s:%s blocked=%s", id, st, reason))
		} else {
			parts = append(parts, fmt.Sprintf("%s:%s", id, st))
		}
	}
	return fmt.Sprintf("staging {%s}", joinParts(parts))
}
