package compressor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/model"
)

// RuntimeStats tracks cumulative compressor run metrics per unit.
type RuntimeStats struct {
	mu         sync.RWMutex
	clk        clock.Clock
	startTimes map[model.CompressorID]time.Time
	runSeconds map[model.CompressorID]float64
	startCount map[model.CompressorID]int
}

func NewRuntimeStats(clk clock.Clock) *RuntimeStats {
	return &RuntimeStats{
		clk:        clk,
		startTimes: make(map[model.CompressorID]time.Time),
		runSeconds: make(map[model.CompressorID]float64),
		startCount: make(map[model.CompressorID]int),
	}
}

func (r *RuntimeStats) OnStart(id model.CompressorID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.startTimes[id] = r.clk.Now()
	r.startCount[id]++
}

func (r *RuntimeStats) OnStop(id model.CompressorID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	start, ok := r.startTimes[id]
	if !ok {
		return
	}
	elapsed := r.clk.Now().Sub(start).Seconds()
	r.runSeconds[id] += elapsed
	delete(r.startTimes, id)
}

func (r *RuntimeStats) TotalRunSeconds(id model.CompressorID) float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	total := r.runSeconds[id]
	if start, ok := r.startTimes[id]; ok {
		total += r.clk.Now().Sub(start).Seconds()
	}
	return total
}

func (r *RuntimeStats) StartCount(id model.CompressorID) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.startCount[id]
}

func (r *RuntimeStats) Report() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	parts := make([]string, 0, len(r.runSeconds))
	for id, sec := range r.runSeconds {
		parts = append(parts, fmt.Sprintf("%s:%.1fs starts=%d", id, sec, r.startCount[id]))
	}
	return fmt.Sprintf("compressor_runtime {%s}", joinParts(parts))
}

func joinParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	s := parts[0]
	for i := 1; i < len(parts); i++ {
		s += " " + parts[i]
	}
	return s
}

// LoadBalancer picks the compressor with lowest cumulative runtime for staging.
type LoadBalancer struct {
	stats *RuntimeStats
}

func NewLoadBalancer(stats *RuntimeStats) *LoadBalancer {
	return &LoadBalancer{stats: stats}
}

func (lb *LoadBalancer) Pick(candidates []model.CompressorID) (model.CompressorID, error) {
	if len(candidates) == 0 {
		return "", model.Wrap("loadbalancer", "empty", model.ErrNotFound)
	}
	best := candidates[0]
	bestRun := lb.stats.TotalRunSeconds(best)
	for _, id := range candidates[1:] {
		run := lb.stats.TotalRunSeconds(id)
		if run < bestRun {
			best = id
			bestRun = run
		}
	}
	return best, nil
}

func (lb *LoadBalancer) StageNext(ctx context.Context, co *Coordinator, candidates []model.CompressorID) error {
	id, err := lb.Pick(candidates)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return model.Wrap("loadbalancer", "canceled", context.Cause(ctx))
	default:
	}
	return co.Start(ctx, id)
}
