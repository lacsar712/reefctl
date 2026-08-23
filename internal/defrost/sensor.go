package defrost

import (
	"sync"

	"github.com/lacsar712/reefctl/internal/clock"
	"github.com/lacsar712/reefctl/internal/model"
)

type SensorBank struct {
	mu   sync.RWMutex
	clk  clock.Clock
	data map[model.SensorID]float64
}

func NewSensorBank(clk clock.Clock) *SensorBank {
	return &SensorBank{clk: clk, data: make(map[model.SensorID]float64)}
}

func (b *SensorBank) Set(id model.SensorID, celsius float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data[id] = celsius
}

func (b *SensorBank) Reading(id model.SensorID) (model.TempReading, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	v, ok := b.data[id]
	if !ok {
		return model.TempReading{}, false
	}
	return model.TempReading{Sensor: id, Celsius: v, At: b.clk.Now()}, true
}

func (b *SensorBank) Average(ids []model.SensorID) (float64, int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var sum float64
	var n int
	for _, id := range ids {
		if v, ok := b.data[id]; ok {
			sum += v
			n++
		}
	}
	if n == 0 {
		return 0, 0
	}
	return sum / float64(n), n
}