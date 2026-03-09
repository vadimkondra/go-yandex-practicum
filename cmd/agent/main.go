package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"go-yandex-practicum/internal/config"
	"go-yandex-practicum/internal/model"
)

var AppConfig config.AgentConfig

var pollCount int64

func main() {
	parseFlags()

	client := &http.Client{}

	var metrics []models.Metrics

	pollTicker := time.NewTicker(time.Duration(AppConfig.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(AppConfig.ReportInterval) * time.Second)

	for {
		select {
		case <-pollTicker.C:
			// обновляем метрики runtime
			metrics = fillMetrics()

		case <-reportTicker.C:
			// отправляем метрики на сервер
			if pollCount > 0 {
				sendMetrics(client, metrics)
				pollCount = 0
			}
		}
	}
}

func parseFlags() {
	flag.StringVar(&AppConfig.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&AppConfig.PollInterval, "p", 2, "polling interval for collecting metrics")
	flag.IntVar(&AppConfig.ReportInterval, "r", 10, "reporting interval for sending metrics to server")

	flag.Parse()
}

func buildUpdateMetricURL(metricType string, metricNm string, metricVal string) string {
	return "update/" + metricType + "/" + metricNm + "/" + metricVal
}

func sendRequest(client *http.Client, url string) {
	req, err := http.NewRequest(http.MethodPost, url, nil)

	if err != nil {
		panic(err)
	}

	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	fmt.Println("url:", url)
	fmt.Println("Status:", response.Status)
}

func sendMetrics(client *http.Client, metrics []models.Metrics) {
	for _, metric := range metrics {
		sendMetric(client, metric)
	}
}

func sendMetric(client *http.Client, metric models.Metrics) {
	var metricValue string

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return
		}
		metricValue = strconv.FormatFloat(*metric.Value, 'f', -1, 64)

	case models.Counter:
		if metric.Delta == nil {
			return
		}
		metricValue = strconv.FormatInt(*metric.Delta, 10)

	default:
		return
	}

	url := "http://" + AppConfig.ServerAddress + "/" + buildUpdateMetricURL(metric.MType, metric.ID, metricValue)

	sendRequest(client, url)
}

func fillMetrics() []models.Metrics {
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	pollCount++

	var metrics []models.Metrics

	metrics = append(metrics,
		models.Metrics{ID: "Alloc", MType: models.Gauge, Value: float64Ptr(float64(memStats.Alloc))},
		models.Metrics{ID: "BuckHashSys", MType: models.Gauge, Value: float64Ptr(float64(memStats.BuckHashSys))},
		models.Metrics{ID: "Frees", MType: models.Gauge, Value: float64Ptr(float64(memStats.Frees))},
		models.Metrics{ID: "GCCPUFraction", MType: models.Gauge, Value: float64Ptr(memStats.GCCPUFraction)},
		models.Metrics{ID: "GCSys", MType: models.Gauge, Value: float64Ptr(float64(memStats.GCSys))},
		models.Metrics{ID: "HeapAlloc", MType: models.Gauge, Value: float64Ptr(float64(memStats.HeapAlloc))},
		models.Metrics{ID: "HeapIdle", MType: models.Gauge, Value: float64Ptr(float64(memStats.HeapIdle))},
		models.Metrics{ID: "HeapInuse", MType: models.Gauge, Value: float64Ptr(float64(memStats.HeapInuse))},
		models.Metrics{ID: "HeapObjects", MType: models.Gauge, Value: float64Ptr(float64(memStats.HeapObjects))},
		models.Metrics{ID: "HeapReleased", MType: models.Gauge, Value: float64Ptr(float64(memStats.HeapReleased))},
		models.Metrics{ID: "HeapSys", MType: models.Gauge, Value: float64Ptr(float64(memStats.HeapSys))},
		models.Metrics{ID: "LastGC", MType: models.Gauge, Value: float64Ptr(float64(memStats.LastGC))},
		models.Metrics{ID: "Lookups", MType: models.Gauge, Value: float64Ptr(float64(memStats.Lookups))},
		models.Metrics{ID: "MCacheInuse", MType: models.Gauge, Value: float64Ptr(float64(memStats.MCacheInuse))},
		models.Metrics{ID: "MCacheSys", MType: models.Gauge, Value: float64Ptr(float64(memStats.MCacheSys))},
		models.Metrics{ID: "MSpanInuse", MType: models.Gauge, Value: float64Ptr(float64(memStats.MSpanInuse))},
		models.Metrics{ID: "MSpanSys", MType: models.Gauge, Value: float64Ptr(float64(memStats.MSpanSys))},
		models.Metrics{ID: "Mallocs", MType: models.Gauge, Value: float64Ptr(float64(memStats.Mallocs))},
		models.Metrics{ID: "NextGC", MType: models.Gauge, Value: float64Ptr(float64(memStats.NextGC))},
		models.Metrics{ID: "NumForcedGC", MType: models.Gauge, Value: float64Ptr(float64(memStats.NumForcedGC))},
		models.Metrics{ID: "NumGC", MType: models.Gauge, Value: float64Ptr(float64(memStats.NumGC))},
		models.Metrics{ID: "OtherSys", MType: models.Gauge, Value: float64Ptr(float64(memStats.OtherSys))},
		models.Metrics{ID: "PauseTotalNs", MType: models.Gauge, Value: float64Ptr(float64(memStats.PauseTotalNs))},
		models.Metrics{ID: "StackInuse", MType: models.Gauge, Value: float64Ptr(float64(memStats.StackInuse))},
		models.Metrics{ID: "StackSys", MType: models.Gauge, Value: float64Ptr(float64(memStats.StackSys))},
		models.Metrics{ID: "Sys", MType: models.Gauge, Value: float64Ptr(float64(memStats.Sys))},
		models.Metrics{ID: "TotalAlloc", MType: models.Gauge, Value: float64Ptr(float64(memStats.TotalAlloc))},
		models.Metrics{ID: "RandomValue", MType: models.Gauge, Value: float64Ptr(rand.Float64())},
		models.Metrics{ID: "PollCount", MType: models.Counter, Delta: int64Ptr(pollCount)},
	)

	return metrics
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
