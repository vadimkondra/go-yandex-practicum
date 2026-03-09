package main

import (
	models "go-yandex-practicum/internal/model"
	"testing"
)

func TestReadMemStatMetrics_AllMetricsPresent(t *testing.T) {
	var metrics []models.Metrics

	metrics = fillMetrics()

	expected := []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"RandomValue",
		"PollCount",
	}

	found := make(map[string]bool)

	for _, m := range metrics {
		found[m.ID] = true
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("metric %s not found in slice", name)
		}
	}
}

func TestBuildUpdateMetricURL(t *testing.T) {
	got := buildUpdateMetricURL("gauge", "Alloc", "123")
	want := "update/gauge/Alloc/123"

	if got != want {
		t.Fatalf("wrong url:\nwant: %s\ngot:  %s", want, got)
	}
}

func TestReadMemStatMetrics_MapIsNotEmpty(t *testing.T) {
	var metrics []models.Metrics

	metrics = fillMetrics()

	if len(metrics) == 0 {
		t.Fatal("metrics map is empty")
	}
}
