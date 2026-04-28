package store

import "testing"

func TestMemStorageSetAndGetGauge(t *testing.T) {
	storage := NewMemoryStorage()

	err := storage.SetGauge("Alloc", 123.45)
	if err != nil {
		t.Fatalf("SetGauge() error = %v", err)
	}

	got, ok, err := storage.GetGauge("Alloc")
	if err != nil {
		t.Fatalf("GetGauge() error = %v", err)
	}

	if !ok {
		t.Fatal("expected gauge to exist")
	}

	if got != 123.45 {
		t.Errorf("expected gauge value 123.45, got %v", got)
	}
}

func TestMemStorageSetGaugeOverridesValue(t *testing.T) {
	storage := NewMemoryStorage()

	if err := storage.SetGauge("Alloc", 100); err != nil {
		t.Fatalf("SetGauge() first error = %v", err)
	}

	if err := storage.SetGauge("Alloc", 200); err != nil {
		t.Fatalf("SetGauge() second error = %v", err)
	}

	got, ok, err := storage.GetGauge("Alloc")
	if err != nil {
		t.Fatalf("GetGauge() error = %v", err)
	}

	if !ok {
		t.Fatal("expected gauge to exist")
	}

	if got != 200 {
		t.Errorf("expected gauge value 200, got %v", got)
	}
}

func TestMemStorageAddAndGetCounter(t *testing.T) {
	storage := NewMemoryStorage()

	got, err := storage.AddCounter("PollCount", 2)
	if err != nil {
		t.Fatalf("AddCounter() first error = %v", err)
	}

	if got != 2 {
		t.Fatalf("expected counter value 2, got %d", got)
	}

	got, err = storage.AddCounter("PollCount", 3)
	if err != nil {
		t.Fatalf("AddCounter() second error = %v", err)
	}

	if got != 5 {
		t.Fatalf("expected counter value 5, got %d", got)
	}

	value, ok, err := storage.GetCounter("PollCount")
	if err != nil {
		t.Fatalf("GetCounter() error = %v", err)
	}

	if !ok {
		t.Fatal("expected counter to exist")
	}

	if value != 5 {
		t.Errorf("expected counter value 5, got %d", value)
	}
}

func TestMemStorageGetUnknownGauge(t *testing.T) {
	storage := NewMemoryStorage()

	_, ok, err := storage.GetGauge("Unknown")
	if err != nil {
		t.Fatalf("GetGauge() error = %v", err)
	}

	if ok {
		t.Fatal("expected gauge to be absent")
	}
}

func TestMemStorageGetUnknownCounter(t *testing.T) {
	storage := NewMemoryStorage()

	_, ok, err := storage.GetCounter("Unknown")
	if err != nil {
		t.Fatalf("GetCounter() error = %v", err)
	}

	if ok {
		t.Fatal("expected counter to be absent")
	}
}

func TestMemStorageGetAllGauges(t *testing.T) {
	storage := NewMemoryStorage()

	if err := storage.SetGauge("Alloc", 123.45); err != nil {
		t.Fatalf("SetGauge() error = %v", err)
	}

	if err := storage.SetGauge("BuckHashSys", 99.9); err != nil {
		t.Fatalf("SetGauge() error = %v", err)
	}

	gauges, err := storage.GetAllGauges()
	if err != nil {
		t.Fatalf("GetAllGauges() error = %v", err)
	}

	if len(gauges) != 2 {
		t.Fatalf("expected 2 gauges, got %d", len(gauges))
	}

	if gauges["Alloc"] != 123.45 {
		t.Errorf("expected Alloc 123.45, got %v", gauges["Alloc"])
	}

	if gauges["BuckHashSys"] != 99.9 {
		t.Errorf("expected BuckHashSys 99.9, got %v", gauges["BuckHashSys"])
	}
}

func TestMemStorageGetAllCounters(t *testing.T) {
	storage := NewMemoryStorage()

	if _, err := storage.AddCounter("PollCount", 2); err != nil {
		t.Fatalf("AddCounter() error = %v", err)
	}

	if _, err := storage.AddCounter("RetryCount", 5); err != nil {
		t.Fatalf("AddCounter() error = %v", err)
	}

	counters, err := storage.GetAllCounters()
	if err != nil {
		t.Fatalf("GetAllCounters() error = %v", err)
	}

	if len(counters) != 2 {
		t.Fatalf("expected 2 counters, got %d", len(counters))
	}

	if counters["PollCount"] != 2 {
		t.Errorf("expected PollCount 2, got %d", counters["PollCount"])
	}

	if counters["RetryCount"] != 5 {
		t.Errorf("expected RetryCount 5, got %d", counters["RetryCount"])
	}
}

func TestMemStoragePing(t *testing.T) {
	storage := NewMemoryStorage()

	if err := storage.Ping(); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
}

func TestMemStorageClose(t *testing.T) {
	storage := NewMemoryStorage()

	if err := storage.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
