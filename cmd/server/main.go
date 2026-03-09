package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

type MetricsStorage interface {
	SetGauge(name string, value float64)
	AddCounter(name string, value int64)

	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
}

const (
	metricTypeRouteName  = "metric-type"
	metricNameRouteName  = "metric-name"
	metricValueRouteName = "metric-value"
)

var storage = MemStorage{
	gauges:   make(map[string]float64),
	counters: make(map[string]int64),
}

func main() {
	r := ConfigServerRouter()

	http.ListenAndServe(":8080", r)
}

func ConfigServerRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/", getMetricsList)

	r.Get("/value/{"+metricTypeRouteName+"}/{"+metricNameRouteName+"}", getMetricValueHandler)

	r.Route("/update", func(r chi.Router) {
		r.Route("/{"+metricTypeRouteName+"}", func(r chi.Router) {
			r.Route("/{"+metricNameRouteName+"}", func(r chi.Router) {
				r.Post("/{"+metricValueRouteName+"}", metricHandler)
			})
		})
	})

	return r
}

func metricHandler(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, metricTypeRouteName)
	metricName := chi.URLParam(r, metricNameRouteName)
	metricValue := chi.URLParam(r, metricValueRouteName)

	if metricName == "" {
		http.Error(rw, "metric name required", http.StatusNotFound)
		return
	}

	switch metricType {
	case "counter":
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(rw, "invalid counter value", http.StatusBadRequest)
			return
		}
		handleCounter(&storage, metricName, val)

	case "gauge":
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, "invalid gauge value", http.StatusBadRequest)
			return
		}
		handleGauge(&storage, metricName, val)

	default:
		http.Error(rw, "unknown metric type", http.StatusBadRequest)
		return
	}
}

func handleGauge(storage *MemStorage, metricName string, metricValue float64) {
	// здесь логика обработки gauge метрики

	storage.SetGauge(metricName, metricValue)
}

func handleCounter(storage *MemStorage, metricName string, metricValue int64) {
	// здесь логика обработки counter метрики

	storage.AddCounter(metricName, metricValue)
}

func getMetricValueHandler(rw http.ResponseWriter, r *http.Request) {
	// тут логика получения значения метрики

	metricType := chi.URLParam(r, metricTypeRouteName)
	metricName := chi.URLParam(r, metricNameRouteName)

	switch metricType {
	case "counter":
		value, ok := storage.GetCounter(metricName)
		if !ok {
			http.Error(rw, "unknown metric name", http.StatusNotFound)
			return
		}

		GetResponseWithMetricValue(rw, strconv.FormatInt(value, 10))
	case "gauge":
		value, ok := storage.GetGauge(metricName)
		if !ok {
			http.Error(rw, "unknown metric name", http.StatusNotFound)
			return
		}

		GetResponseWithMetricValue(rw, strconv.FormatFloat(value, 'f', -1, 64))
	default:
		http.Error(rw, "unknown metric type", http.StatusNotFound)
		return
	}
}

func GetResponseWithMetricValue(rw http.ResponseWriter, metricValue string) {
	// логика получения значения метрики и формирования ответа

	rw.Write([]byte(metricValue))
}

func getMetricsList(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)

	io.WriteString(rw, "<html><body>")
	io.WriteString(rw, "<h1>Metrics</h1>")

	io.WriteString(rw, "<h2>Gauges</h2><ul>")
	for name, value := range storage.gauges {
		io.WriteString(rw, fmt.Sprintf("<li>%s: %v</li>", name, value))
	}
	io.WriteString(rw, "</ul>")

	io.WriteString(rw, "<h2>Counters</h2><ul>")
	for name, value := range storage.counters {
		io.WriteString(rw, fmt.Sprintf("<li>%s: %d</li>", name, value))
	}
	io.WriteString(rw, "</ul>")

	io.WriteString(rw, "</body></html>")
}

func (s *MemStorage) SetGauge(name string, value float64) {
	s.gauges[name] = value
}

func (s *MemStorage) AddCounter(name string, value int64) {
	s.counters[name] += value
}

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	v, ok := s.gauges[name]
	return v, ok
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	v, ok := s.counters[name]
	return v, ok
}
