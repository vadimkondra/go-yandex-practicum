package service

import (
	"encoding/json"
	models "go-yandex-practicum/internal/model"
	"go-yandex-practicum/internal/repository"
	"log"
	"os"
	"time"
)

var storage repository.MetricsStorage = repository.NewMemStorage()

func SetGauge(metricName string, metricValue float64) {
	storage.SetGauge(metricName, metricValue)
}

func AddCounter(metricName string, metricValue int64) int64 {
	return storage.AddCounter(metricName, metricValue)
}

func GetGauge(metricName string) (float64, bool) {
	return storage.GetGauge(metricName)
}

func GetCounter(metricName string) (int64, bool) {
	return storage.GetCounter(metricName)
}

func GetAllGauges() map[string]float64 {
	return storage.GetAllGauges()
}

func GetAllCounters() map[string]int64 {
	return storage.GetAllCounters()
}

func StoreMetrics(storeInterval int, filePath string) {
	ticker := time.NewTicker(time.Duration(storeInterval) * time.Second)

	defer ticker.Stop()
	for range ticker.C {
		if err := saveMetricsToFile(filePath); err != nil {
			log.Printf("saveMetricsToFile error", err)
		}
	}
}

func saveMetricsToFile(filePath string) error {

	metrics := make([]models.Metrics, 0)
	for name, value := range storage.GetAllGauges() {
		v := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &v,
		})
	}
	for name, value := range storage.GetAllCounters() {
		v := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &v,
		})
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(metrics)

}

func LoadMetricsFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var metrics []models.Metrics
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		return err
	}

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				storage.SetGauge(metric.ID, *metric.Value)
			}
		case models.Counter:
			if metric.Delta != nil {
				storage.AddCounter(metric.ID, *metric.Delta)
			}
		}
	}

	return nil
}
