package service

import (
	"go-yandex-practicum/internal/model"
	"go-yandex-practicum/internal/store"
)

type MetricsService struct {
	storage store.Storage
}

func NewMetricsService(storage store.Storage) *MetricsService {
	return &MetricsService{storage: storage}
}

func (s *MetricsService) SetGauge(metricName string, metricValue float64) error {
	return s.storage.SetGauge(metricName, metricValue)
}

func (s *MetricsService) AddCounter(metricName string, metricValue int64) (int64, error) {
	return s.storage.AddCounter(metricName, metricValue)
}

func (s *MetricsService) GetGauge(metricName string) (float64, bool, error) {
	return s.storage.GetGauge(metricName)
}

func (s *MetricsService) GetCounter(metricName string) (int64, bool, error) {
	return s.storage.GetCounter(metricName)
}

func (s *MetricsService) GetAllGauges() (map[string]float64, error) {
	return s.storage.GetAllGauges()
}

func (s *MetricsService) GetAllCounters() (map[string]int64, error) {
	return s.storage.GetAllCounters()
}

func (s *MetricsService) Ping() (bool, error) {
	if err := s.storage.Ping(); err != nil {
		return false, err
	}

	return true, nil
}

func (s *MetricsService) UpdateMetricsBatch(metrics []model.Metrics) ([]model.Metrics, error) {
	return s.storage.UpdateBatch(metrics)
}

var service = NewMetricsService(store.NewMemoryStorage())

func SetStorage(s store.Storage) {
	service = NewMetricsService(s)
}

func SetGauge(metricName string, metricValue float64) error {
	return service.SetGauge(metricName, metricValue)
}

func AddCounter(metricName string, metricValue int64) (int64, error) {
	return service.AddCounter(metricName, metricValue)
}

func GetGauge(metricName string) (float64, bool, error) {
	return service.GetGauge(metricName)
}

func GetCounter(metricName string) (int64, bool, error) {
	return service.GetCounter(metricName)
}

func GetAllGauges() (map[string]float64, error) {
	return service.GetAllGauges()
}

func GetAllCounters() (map[string]int64, error) {
	return service.GetAllCounters()
}

func Ping() (bool, error) {
	return service.Ping()
}

func UpdateMetricsBatch(metrics []model.Metrics) ([]model.Metrics, error) {
	return service.UpdateMetricsBatch(metrics)
}
