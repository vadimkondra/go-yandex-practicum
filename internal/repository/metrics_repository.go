package repository

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

type MetricsStorage interface {
	SetGauge(name string, value float64)
	AddCounter(name string, value int64) int64

	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)

	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorage) SetGauge(name string, value float64) {
	s.gauges[name] = value
}

func (s *MemStorage) AddCounter(name string, value int64) int64 {
	s.counters[name] += value
	return s.counters[name]
}

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	v, ok := s.gauges[name]
	return v, ok
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	v, ok := s.counters[name]
	return v, ok
}

func (s *MemStorage) GetAllGauges() map[string]float64 {
	return s.gauges
}

func (s *MemStorage) GetAllCounters() map[string]int64 {
	return s.counters
}
