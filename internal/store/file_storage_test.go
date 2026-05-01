package store

import (
	"encoding/json"
	"go-yandex-practicum/internal/model"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStorageSetGaugeSavesMetricWhenStoreIntervalIsZero(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "metrics.json")

	storage, err := NewFileStorage(filePath, 0, false)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v", err)
	}
	defer storage.Close()

	err = storage.SetGauge("Alloc", 123.45)
	if err != nil {
		t.Fatalf("SetGauge() error = %v", err)
	}

	metrics := readMetricsFromFile(t, filePath)

	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	got := metrics[0]

	if got.ID != "Alloc" {
		t.Errorf("expected ID Alloc, got %s", got.ID)
	}

	if got.MType != model.Gauge {
		t.Errorf("expected type gauge, got %s", got.MType)
	}

	if got.Value == nil {
		t.Fatal("expected gauge value, got nil")
	}

	if *got.Value != 123.45 {
		t.Errorf("expected value 123.45, got %v", *got.Value)
	}
}

func TestFileStorageAddCounterSavesMetricWhenStoreIntervalIsZero(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "metrics.json")

	storage, err := NewFileStorage(filePath, 0, false)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v", err)
	}
	defer storage.Close()

	value, err := storage.AddCounter("PollCount", 2)
	if err != nil {
		t.Fatalf("AddCounter() error = %v", err)
	}

	if value != 2 {
		t.Fatalf("expected counter value 2, got %d", value)
	}

	value, err = storage.AddCounter("PollCount", 3)
	if err != nil {
		t.Fatalf("AddCounter() second call error = %v", err)
	}

	if value != 5 {
		t.Fatalf("expected counter value 5, got %d", value)
	}

	metrics := readMetricsFromFile(t, filePath)

	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	got := metrics[0]

	if got.ID != "PollCount" {
		t.Errorf("expected ID PollCount, got %s", got.ID)
	}

	if got.MType != model.Counter {
		t.Errorf("expected type counter, got %s", got.MType)
	}

	if got.Delta == nil {
		t.Fatal("expected counter delta, got nil")
	}

	if *got.Delta != 5 {
		t.Errorf("expected delta 5, got %d", *got.Delta)
	}
}

func TestFileStorageRestoreLoadsMetricsFromFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "metrics.json")

	gaugeValue := 10.5
	counterValue := int64(7)

	metrics := []model.Metrics{
		{
			ID:    "Alloc",
			MType: model.Gauge,
			Value: &gaugeValue,
		},
		{
			ID:    "PollCount",
			MType: model.Counter,
			Delta: &counterValue,
		},
	}

	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("create file error = %v", err)
	}

	if err := json.NewEncoder(file).Encode(metrics); err != nil {
		file.Close()
		t.Fatalf("encode metrics error = %v", err)
	}

	if err := file.Close(); err != nil {
		t.Fatalf("close file error = %v", err)
	}

	storage, err := NewFileStorage(filePath, 0, true)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v", err)
	}
	defer storage.Close()

	gotGauge, ok, err := storage.GetGauge("Alloc")
	if err != nil {
		t.Fatalf("GetGauge() error = %v", err)
	}

	if !ok {
		t.Fatal("expected gauge to exist")
	}

	if gotGauge != gaugeValue {
		t.Errorf("expected gauge %v, got %v", gaugeValue, gotGauge)
	}

	gotCounter, ok, err := storage.GetCounter("PollCount")
	if err != nil {
		t.Fatalf("GetCounter() error = %v", err)
	}

	if !ok {
		t.Fatal("expected counter to exist")
	}

	if gotCounter != counterValue {
		t.Errorf("expected counter %d, got %d", counterValue, gotCounter)
	}
}

func TestFileStorageRestoreIgnoresMissingFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "missing.json")

	storage, err := NewFileStorage(filePath, 0, true)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v", err)
	}
	defer storage.Close()

	_, ok, err := storage.GetGauge("Alloc")
	if err != nil {
		t.Fatalf("GetGauge() error = %v", err)
	}

	if ok {
		t.Fatal("expected gauge to be absent")
	}
}

func TestFileStorageCloseSavesMetricsWhenStoreIntervalIsPositive(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "metrics.json")

	storage, err := NewFileStorage(filePath, 1000, false)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v", err)
	}

	err = storage.SetGauge("Alloc", 99.9)
	if err != nil {
		t.Fatalf("SetGauge() error = %v", err)
	}

	err = storage.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	metrics := readMetricsFromFile(t, filePath)

	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	got := metrics[0]

	if got.ID != "Alloc" {
		t.Errorf("expected ID Alloc, got %s", got.ID)
	}

	if got.Value == nil {
		t.Fatal("expected gauge value, got nil")
	}

	if *got.Value != 99.9 {
		t.Errorf("expected value 99.9, got %v", *got.Value)
	}
}

func readMetricsFromFile(t *testing.T, filePath string) []model.Metrics {
	t.Helper()

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("open file error = %v", err)
	}
	defer file.Close()

	var metrics []model.Metrics

	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		t.Fatalf("decode metrics error = %v", err)
	}

	return metrics
}
