package main

import "testing"

func TestReadMemStatMetrics_AllMetricsPresent(t *testing.T) {
	metrics := make(map[string]gauge)

	readMemStatMetrics(metrics)

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
	}

	for _, name := range expected {
		if _, ok := metrics[name]; !ok {
			t.Errorf("metric %s not found in map", name)
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
	metrics := make(map[string]gauge)

	readMemStatMetrics(metrics)

	if len(metrics) == 0 {
		t.Fatal("metrics map is empty")
	}
}
