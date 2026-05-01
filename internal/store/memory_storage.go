package store

import (
	"go-yandex-practicum/internal/model"
	"sync"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}

func NewMemoryStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorage) SetGauge(name string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
	return nil
}

func (s *MemStorage) AddCounter(name string, value int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += value
	return s.counters[name], nil
}

func (s *MemStorage) GetGauge(name string) (float64, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.gauges[name]
	return v, ok, nil
}

func (s *MemStorage) GetCounter(name string) (int64, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.counters[name]
	return v, ok, nil
}

func (s *MemStorage) GetAllGauges() (map[string]float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	gauges := make(map[string]float64, len(s.gauges))
	for name, value := range s.gauges {
		gauges[name] = value
	}

	return gauges, nil
}

func (s *MemStorage) GetAllCounters() (map[string]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counters := make(map[string]int64, len(s.counters))
	for name, value := range s.counters {
		counters[name] = value
	}

	return counters, nil
}

func (s *MemStorage) Ping() error {
	return nil
}

func (s *MemStorage) Close() error {
	return nil
}

func (s *MemStorage) UpdateBatch(metrics []model.Metrics) ([]model.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	updated := make([]model.Metrics, 0, len(metrics))
	for _, metric := range metrics {
		switch metric.MType {
		case model.Gauge:
			if metric.Value == nil {
				continue
			}
			s.gauges[metric.ID] = *metric.Value
			value := s.gauges[metric.ID]
			updated = append(updated, model.Metrics{
				ID:    metric.ID,
				MType: model.Gauge,
				Value: &value,
			})
		case model.Counter:
			if metric.Delta == nil {
				continue
			}
			s.counters[metric.ID] += *metric.Delta
			delta := s.counters[metric.ID]
			updated = append(updated, model.Metrics{
				ID:    metric.ID,
				MType: model.Counter,
				Delta: &delta,
			})
		}
	}
	return updated, nil
}
