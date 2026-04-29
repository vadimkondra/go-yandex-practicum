package service

import (
	"errors"
	"go-yandex-practicum/internal/model"
	"testing"
)

type mockStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	err      error
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func setTestStorage(t *testing.T) *mockStorage {
	t.Helper()

	mock := newMockStorage()
	SetStorage(mock)

	return mock
}

func (m *mockStorage) SetGauge(name string, value float64) error {
	if m.err != nil {
		return m.err
	}

	m.gauges[name] = value
	return nil
}

func (m *mockStorage) AddCounter(name string, delta int64) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}

	m.counters[name] += delta
	return m.counters[name], nil
}

func (m *mockStorage) GetGauge(name string) (float64, bool, error) {
	if m.err != nil {
		return 0, false, m.err
	}

	value, ok := m.gauges[name]
	return value, ok, nil
}

func (m *mockStorage) GetCounter(name string) (int64, bool, error) {
	if m.err != nil {
		return 0, false, m.err
	}

	value, ok := m.counters[name]
	return value, ok, nil
}

func (m *mockStorage) GetAllGauges() (map[string]float64, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.gauges, nil
}

func (m *mockStorage) GetAllCounters() (map[string]int64, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.counters, nil
}

func (m *mockStorage) Ping() error {
	return m.err
}

func (m *mockStorage) Close() error {
	return nil
}

func (m *mockStorage) UpdateBatch(metrics []model.Metrics) ([]model.Metrics, error) {
	if m.err != nil {
		return nil, m.err
	}

	updated := make([]model.Metrics, 0, len(metrics))

	for _, metric := range metrics {
		switch metric.MType {
		case model.Gauge:
			if metric.Value == nil {
				continue
			}

			m.gauges[metric.ID] = *metric.Value

			value := m.gauges[metric.ID]
			updated = append(updated, model.Metrics{
				ID:    metric.ID,
				MType: model.Gauge,
				Value: &value,
			})

		case model.Counter:
			if metric.Delta == nil {
				continue
			}

			m.counters[metric.ID] += *metric.Delta

			delta := m.counters[metric.ID]
			updated = append(updated, model.Metrics{
				ID:    metric.ID,
				MType: model.Counter,
				Delta: &delta,
			})
		}
	}

	return updated, nil
}

func TestServiceSetAndGetGauge(t *testing.T) {
	setTestStorage(t)

	err := SetGauge("Alloc", 123.45)
	if err != nil {
		t.Fatalf("SetGauge() error = %v", err)
	}

	got, ok, err := GetGauge("Alloc")
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

func TestServiceAddAndGetCounter(t *testing.T) {
	setTestStorage(t)

	got, err := AddCounter("PollCount", 2)
	if err != nil {
		t.Fatalf("AddCounter() first error = %v", err)
	}

	if got != 2 {
		t.Fatalf("expected counter value 2, got %d", got)
	}

	got, err = AddCounter("PollCount", 3)
	if err != nil {
		t.Fatalf("AddCounter() second error = %v", err)
	}

	if got != 5 {
		t.Fatalf("expected counter value 5, got %d", got)
	}

	value, ok, err := GetCounter("PollCount")
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

func TestServiceGetAllGauges(t *testing.T) {
	setTestStorage(t)

	if err := SetGauge("Alloc", 1.5); err != nil {
		t.Fatalf("SetGauge() error = %v", err)
	}

	if err := SetGauge("BuckHashSys", 2.5); err != nil {
		t.Fatalf("SetGauge() error = %v", err)
	}

	gauges, err := GetAllGauges()
	if err != nil {
		t.Fatalf("GetAllGauges() error = %v", err)
	}

	if len(gauges) != 2 {
		t.Fatalf("expected 2 gauges, got %d", len(gauges))
	}

	if gauges["Alloc"] != 1.5 {
		t.Errorf("expected Alloc 1.5, got %v", gauges["Alloc"])
	}

	if gauges["BuckHashSys"] != 2.5 {
		t.Errorf("expected BuckHashSys 2.5, got %v", gauges["BuckHashSys"])
	}
}

func TestServiceGetAllCounters(t *testing.T) {
	setTestStorage(t)

	if _, err := AddCounter("PollCount", 2); err != nil {
		t.Fatalf("AddCounter() error = %v", err)
	}

	if _, err := AddCounter("RetryCount", 5); err != nil {
		t.Fatalf("AddCounter() error = %v", err)
	}

	counters, err := GetAllCounters()
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

func TestServicePingSuccess(t *testing.T) {
	setTestStorage(t)

	ok, err := Ping()
	if err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	if !ok {
		t.Fatal("expected ping result true")
	}
}

func TestServiceReturnsStorageError(t *testing.T) {
	expectedErr := errors.New("storage error")

	mock := setTestStorage(t)
	mock.err = expectedErr

	if err := SetGauge("Alloc", 1); !errors.Is(err, expectedErr) {
		t.Fatalf("expected SetGauge error %v, got %v", expectedErr, err)
	}

	if _, err := AddCounter("PollCount", 1); !errors.Is(err, expectedErr) {
		t.Fatalf("expected AddCounter error %v, got %v", expectedErr, err)
	}

	if _, _, err := GetGauge("Alloc"); !errors.Is(err, expectedErr) {
		t.Fatalf("expected GetGauge error %v, got %v", expectedErr, err)
	}

	if _, _, err := GetCounter("PollCount"); !errors.Is(err, expectedErr) {
		t.Fatalf("expected GetCounter error %v, got %v", expectedErr, err)
	}

	if _, err := GetAllGauges(); !errors.Is(err, expectedErr) {
		t.Fatalf("expected GetAllGauges error %v, got %v", expectedErr, err)
	}

	if _, err := GetAllCounters(); !errors.Is(err, expectedErr) {
		t.Fatalf("expected GetAllCounters error %v, got %v", expectedErr, err)
	}

	ok, err := Ping()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected Ping error %v, got %v", expectedErr, err)
	}

	if ok {
		t.Fatal("expected ping result false")
	}
}
