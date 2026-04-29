package store

import (
	"encoding/json"
	"errors"
	"go-yandex-practicum/internal/model"
	"os"
	"time"
)

type FileStorage struct {
	memory        *MemStorage
	filePath      string
	storeInterval int
	done          chan struct{}
}

func NewFileStorage(filePath string, storeInterval int, restore bool) (*FileStorage, error) {

	storage := &FileStorage{
		memory:        NewMemoryStorage(),
		filePath:      filePath,
		storeInterval: storeInterval,
		done:          make(chan struct{}),
	}
	if restore {
		if err := storage.load(); err != nil {
			return nil, err
		}
	}
	if storeInterval > 0 {
		go storage.storePeriodically()
	}
	return storage, nil
}

func (s *FileStorage) SetGauge(name string, value float64) error {

	if err := s.memory.SetGauge(name, value); err != nil {
		return err
	}
	return s.saveSyncIfNeeded()
}

func (s *FileStorage) AddCounter(name string, delta int64) (int64, error) {

	value, err := s.memory.AddCounter(name, delta)
	if err != nil {
		return 0, err
	}
	if err := s.saveSyncIfNeeded(); err != nil {
		return 0, err
	}
	return value, nil
}

func (s *FileStorage) GetGauge(name string) (float64, bool, error) {
	return s.memory.GetGauge(name)
}

func (s *FileStorage) GetCounter(name string) (int64, bool, error) {
	return s.memory.GetCounter(name)
}

func (s *FileStorage) GetAllGauges() (map[string]float64, error) {
	return s.memory.GetAllGauges()
}

func (s *FileStorage) GetAllCounters() (map[string]int64, error) {
	return s.memory.GetAllCounters()
}

func (s *FileStorage) Ping() error {
	return nil
}

func (s *FileStorage) Close() error {

	close(s.done)
	if s.storeInterval > 0 {
		return s.save()
	}
	return nil
}

func (s *FileStorage) saveSyncIfNeeded() error {

	if s.storeInterval == 0 {
		return s.save()
	}
	return nil
}

func (s *FileStorage) storePeriodically() {

	ticker := time.NewTicker(time.Duration(s.storeInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = s.save()
		case <-s.done:
			return
		}
	}
}

func (s *FileStorage) save() error {

	metrics := make([]model.Metrics, 0)
	gauges, err := s.memory.GetAllGauges()
	if err != nil {
		return err
	}
	for name, value := range gauges {
		v := value
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: model.Gauge,
			Value: &v,
		})
	}
	counters, err := s.memory.GetAllCounters()
	if err != nil {
		return err
	}
	for name, value := range counters {
		v := value
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: model.Counter,
			Delta: &v,
		})
	}
	file, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(metrics)
}

func (s *FileStorage) load() error {

	file, err := os.Open(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()
	var metrics []model.Metrics
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		return err
	}
	for _, metric := range metrics {
		switch metric.MType {
		case model.Gauge:
			if metric.Value != nil {
				if err := s.memory.SetGauge(metric.ID, *metric.Value); err != nil {
					return err
				}
			}
		case model.Counter:
			if metric.Delta != nil {
				if _, err := s.memory.AddCounter(metric.ID, *metric.Delta); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *FileStorage) UpdateBatch(metrics []model.Metrics) ([]model.Metrics, error) {

	updated, err := s.memory.UpdateBatch(metrics)
	if err != nil {
		return nil, err
	}
	if err := s.saveSyncIfNeeded(); err != nil {
		return nil, err
	}
	return updated, nil

}
