package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
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

	var metrics []model.Metrics

	pollTicker := time.NewTicker(time.Duration(AppConfig.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(AppConfig.ReportInterval) * time.Second)

	for {
		select {
		case <-pollTicker.C:
			// обновляем метрики runtime
			metrics = fillMetrics()

		case <-reportTicker.C:
			// отправляем метрики на сервер батчем
			if len(metrics) > 0 {
				err := sendMetricsBatch(client, metrics)
				if err != nil {
					log.Println("send metrics batch error:", err)
				} else {
					pollCount = 0
					metrics = nil
				}
			}
		}
	}
}

func parseFlags() {
	flag.StringVar(&AppConfig.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&AppConfig.PollInterval, "p", 2, "polling interval for collecting metrics")
	flag.IntVar(&AppConfig.ReportInterval, "r", 10, "reporting interval for sending metrics to server")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		AppConfig.ServerAddress = envRunAddr
	}
	if envRunReportInterval := os.Getenv("REPORT_INTERVAL"); envRunReportInterval != "" {
		value, err := strconv.Atoi(envRunReportInterval)
		if err != nil {
			log.Fatal("invalid REPORT_INTERVAL:", err)
		}

		AppConfig.ReportInterval = value
	}
	if envRunPoolInterval := os.Getenv("POLL_INTERVAL"); envRunPoolInterval != "" {
		value, err := strconv.Atoi(envRunPoolInterval)
		if err != nil {
			log.Fatal("invalid POLL_INTERVAL:", err)
		}

		AppConfig.PollInterval = value
	}
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

func sendMetricsBatch(client *http.Client, metrics []model.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics batch: %w", err)
	}

	url := "http://" + AppConfig.ServerAddress + "/updates/"

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
