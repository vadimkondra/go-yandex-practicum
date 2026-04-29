package store

import models "go-yandex-practicum/internal/model"

type Storage interface {
	SetGauge(name string, value float64) error
	AddCounter(name string, delta int64) (int64, error)

	UpdateBatch(metrics []models.Metrics) ([]models.Metrics, error)

	GetGauge(name string) (float64, bool, error)
	GetCounter(name string) (int64, bool, error)
	GetAllGauges() (map[string]float64, error)
	GetAllCounters() (map[string]int64, error)
	Ping() error
	Close() error
}
