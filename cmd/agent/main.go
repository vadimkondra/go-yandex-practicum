package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"go-yandex-practicum/internal/retry"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"go-yandex-practicum/internal/config"
	"go-yandex-practicum/internal/model"
)

var AppConfig config.AgentConfig

var pollCount int64

func main() {
	ParseFlags()

	client := &http.Client{}

	var metrics []model.Metrics

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
				err := sendMetrics(client, metrics)
				if err != nil {
					log.Println("send metrics error:", err)
				} else {
					pollCount = 0
				}
			}
		}
	}
}

func sendRequest(client *http.Client, url string, body []byte) error {

	return retry.Do(func() error {
		return sendRequestOnce(client, url, body)
	}, isRetriableHTTPError)

}

func isRetriableHTTPError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	return false
}

func sendRequestOnce(client *http.Client, url string, body []byte) error {
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

func sendMetrics(client *http.Client, metrics []model.Metrics) error {
	for _, metric := range metrics {
		if err := sendMetric(client, metric); err != nil {
			return err
		}
	}

	return nil
}

func sendMetric(client *http.Client, metric model.Metrics) error {
	switch metric.MType {
	case model.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("gauge metric %q has nil value", metric.ID)
		}

	case model.Counter:
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

	url := "http://" + AppConfig.ServerAddress + "/update/"

	return sendRequest(client, url, body)
}

func fillMetrics() []model.Metrics {
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	pollCount++

	var metrics []model.Metrics

	metrics = append(metrics,
		model.Metrics{ID: "Alloc", MType: model.Gauge, Value: float64Ptr(float64(memStats.Alloc))},
		model.Metrics{ID: "BuckHashSys", MType: model.Gauge, Value: float64Ptr(float64(memStats.BuckHashSys))},
		model.Metrics{ID: "Frees", MType: model.Gauge, Value: float64Ptr(float64(memStats.Frees))},
		model.Metrics{ID: "GCCPUFraction", MType: model.Gauge, Value: float64Ptr(memStats.GCCPUFraction)},
		model.Metrics{ID: "GCSys", MType: model.Gauge, Value: float64Ptr(float64(memStats.GCSys))},
		model.Metrics{ID: "HeapAlloc", MType: model.Gauge, Value: float64Ptr(float64(memStats.HeapAlloc))},
		model.Metrics{ID: "HeapIdle", MType: model.Gauge, Value: float64Ptr(float64(memStats.HeapIdle))},
		model.Metrics{ID: "HeapInuse", MType: model.Gauge, Value: float64Ptr(float64(memStats.HeapInuse))},
		model.Metrics{ID: "HeapObjects", MType: model.Gauge, Value: float64Ptr(float64(memStats.HeapObjects))},
		model.Metrics{ID: "HeapReleased", MType: model.Gauge, Value: float64Ptr(float64(memStats.HeapReleased))},
		model.Metrics{ID: "HeapSys", MType: model.Gauge, Value: float64Ptr(float64(memStats.HeapSys))},
		model.Metrics{ID: "LastGC", MType: model.Gauge, Value: float64Ptr(float64(memStats.LastGC))},
		model.Metrics{ID: "Lookups", MType: model.Gauge, Value: float64Ptr(float64(memStats.Lookups))},
		model.Metrics{ID: "MCacheInuse", MType: model.Gauge, Value: float64Ptr(float64(memStats.MCacheInuse))},
		model.Metrics{ID: "MCacheSys", MType: model.Gauge, Value: float64Ptr(float64(memStats.MCacheSys))},
		model.Metrics{ID: "MSpanInuse", MType: model.Gauge, Value: float64Ptr(float64(memStats.MSpanInuse))},
		model.Metrics{ID: "MSpanSys", MType: model.Gauge, Value: float64Ptr(float64(memStats.MSpanSys))},
		model.Metrics{ID: "Mallocs", MType: model.Gauge, Value: float64Ptr(float64(memStats.Mallocs))},
		model.Metrics{ID: "NextGC", MType: model.Gauge, Value: float64Ptr(float64(memStats.NextGC))},
		model.Metrics{ID: "NumForcedGC", MType: model.Gauge, Value: float64Ptr(float64(memStats.NumForcedGC))},
		model.Metrics{ID: "NumGC", MType: model.Gauge, Value: float64Ptr(float64(memStats.NumGC))},
		model.Metrics{ID: "OtherSys", MType: model.Gauge, Value: float64Ptr(float64(memStats.OtherSys))},
		model.Metrics{ID: "PauseTotalNs", MType: model.Gauge, Value: float64Ptr(float64(memStats.PauseTotalNs))},
		model.Metrics{ID: "StackInuse", MType: model.Gauge, Value: float64Ptr(float64(memStats.StackInuse))},
		model.Metrics{ID: "StackSys", MType: model.Gauge, Value: float64Ptr(float64(memStats.StackSys))},
		model.Metrics{ID: "Sys", MType: model.Gauge, Value: float64Ptr(float64(memStats.Sys))},
		model.Metrics{ID: "TotalAlloc", MType: model.Gauge, Value: float64Ptr(float64(memStats.TotalAlloc))},
		model.Metrics{ID: "RandomValue", MType: model.Gauge, Value: float64Ptr(rand.Float64())},
		model.Metrics{ID: "PollCount", MType: model.Counter, Delta: int64Ptr(pollCount)},
	)

	return metrics
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
