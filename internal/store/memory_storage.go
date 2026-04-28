package store

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemoryStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorage) SetGauge(name string, value float64) error {
	s.gauges[name] = value
	return nil
}

func (s *MemStorage) AddCounter(name string, value int64) (int64, error) {
	s.counters[name] += value
	return s.counters[name], nil
}

func (s *MemStorage) GetGauge(name string) (float64, bool, error) {
	v, ok := s.gauges[name]
	return v, ok, nil
}

func (s *MemStorage) GetCounter(name string) (int64, bool, error) {
	v, ok := s.counters[name]
	return v, ok, nil
}

func (s *MemStorage) GetAllGauges() (map[string]float64, error) {
	return s.gauges, nil
}

func (s *MemStorage) GetAllCounters() (map[string]int64, error) {
	return s.counters, nil
}

func (s *MemStorage) Ping() error {
	return nil
}

func (s *MemStorage) Close() error {
	return nil
}
