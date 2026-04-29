package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"go-yandex-practicum/internal/config"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"go-yandex-practicum/internal/model"
)

var pollCount int64

func main() {
	cfg := ParseFlags()

	client := &http.Client{}

	var metrics []models.Metrics

	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)

	for {
		select {
		case <-pollTicker.C:
			// обновляем метрики runtime
			metrics = fillMetrics()

		case <-reportTicker.C:
			// отправляем метрики на сервер
			if pollCount > 0 {
				err := sendMetrics(cfg, client, metrics)
				if err != nil {
					log.Println("send metrics error:", err)
				} else {
					pollCount = 0
				}
			}
		}
	}
}

func buildUpdateMetricURL(metricType string, metricNm string, metricVal string) string {
	return "update/" + metricType + "/" + metricNm + "/" + metricVal
}

func sendRequest(client *http.Client, url string, body []byte) error {
	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(body); err != nil {
		return fmt.Errorf("gzip write request body: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("gzip close request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer response.Body.Close()

	var responseBody io.Reader = response.Body
	if response.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(response.Body)
		if err != nil {
			return fmt.Errorf("gzip read response body: %w", err)
		}
		defer gzReader.Close()

		responseBody = gzReader
	}

	_, err = io.ReadAll(responseBody)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}

func sendMetrics(cfg config.AgentConfig, client *http.Client, metrics []models.Metrics) error {
	for _, metric := range metrics {
		if err := sendMetric(cfg, client, metric); err != nil {
			return err
		}
	}

	return nil
}

func sendMetric(cfg config.AgentConfig, client *http.Client, metric models.Metrics) error {
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("gauge metric %q has nil value", metric.ID)
		}

	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("counter metric %q has nil delta", metric.ID)
		}

	default:
		return fmt.Errorf("unknown metric type %q", metric.MType)
	}

	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("marshal metric: %w", err)
	}

	url := "http://" + cfg.ServerAddress + "/update/"

	return sendRequest(client, url, body)
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
