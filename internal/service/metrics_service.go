package service

import (
	"go-yandex-practicum/internal/model"
	"go-yandex-practicum/internal/store"
)

var storage store.Storage

func SetStorage(s store.Storage) {
	storage = s
}

func SetGauge(metricName string, metricValue float64) error {
	return storage.SetGauge(metricName, metricValue)
}

func AddCounter(metricName string, metricValue int64) (int64, error) {
	return storage.AddCounter(metricName, metricValue)
}

func GetGauge(metricName string) (float64, bool, error) {
	return storage.GetGauge(metricName)
}

func GetCounter(metricName string) (int64, bool, error) {
	return storage.GetCounter(metricName)
}

func GetAllGauges() (map[string]float64, error) {
	return storage.GetAllGauges()
}

func GetAllCounters() (map[string]int64, error) {
	return storage.GetAllCounters()
}

func Ping() (bool, error) {
	if err := storage.Ping(); err != nil {
		return false, err
	}
	return true, nil
}

func UpdateMetricsBatch(metrics []model.Metrics) ([]model.Metrics, error) {
	return storage.UpdateBatch(metrics)
}
