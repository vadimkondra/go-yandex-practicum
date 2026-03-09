package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"go-yandex-practicum/internal/config"
	"go-yandex-practicum/internal/model"

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

	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

const (
	metricTypeRouteName  = "metric-type"
	metricNameRouteName  = "metric-name"
	metricValueRouteName = "metric-value"
)

var storage MetricsStorage = &MemStorage{
	gauges:   make(map[string]float64),
	counters: make(map[string]int64),
}

var AppConfig config.ServerConfig

func main() {
	parseFlags()

	r := ConfigServerRouter()

	if err := http.ListenAndServe(AppConfig.ServerAddress, r); err != nil {
		panic(err)
	}
}

func parseFlags() {
	flag.StringVar(&AppConfig.ServerAddress, "a", "localhost:8080", "address and port to run server")

	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}

func ConfigServerRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/", getMetricsListHandler)

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
	case models.Counter:
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(rw, "invalid counter value", http.StatusBadRequest)
			return
		}
		addCounter(storage, metricName, val)

	case models.Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, "invalid gauge value", http.StatusBadRequest)
			return
		}
		setGauge(storage, metricName, val)

	default:
		http.Error(rw, "unknown metric type", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func setGauge(storage MetricsStorage, metricName string, metricValue float64) {
	storage.SetGauge(metricName, metricValue)
}

func addCounter(storage MetricsStorage, metricName string, metricValue int64) {
	storage.AddCounter(metricName, metricValue)
}

func getGauge(storage MetricsStorage, metricName string) (float64, bool) {
	return storage.GetGauge(metricName)
}

func getCounter(storage MetricsStorage, metricName string) (int64, bool) {
	return storage.GetCounter(metricName)
}

func getMetricValueHandler(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, metricTypeRouteName)
	metricName := chi.URLParam(r, metricNameRouteName)

	switch metricType {
	case models.Counter:
		value, ok := getCounter(storage, metricName)
		if !ok {
			http.Error(rw, "unknown metric name", http.StatusNotFound)
			return
		}

		writeMetricValueResponse(rw, strconv.FormatInt(value, 10))
	case models.Gauge:
		value, ok := getGauge(storage, metricName)
		if !ok {
			http.Error(rw, "unknown metric name", http.StatusNotFound)
			return
		}

		writeMetricValueResponse(rw, strconv.FormatFloat(value, 'f', -1, 64))
	default:
		http.Error(rw, "unknown metric type", http.StatusNotFound)
		return
	}
}

func writeMetricValueResponse(rw http.ResponseWriter, metricValue string) {
	rw.Write([]byte(metricValue))
}

func getMetricsListHandler(rw http.ResponseWriter, r *http.Request) {
	buildMetricsListResponse(storage, rw)
}

func buildMetricsListResponse(storage MetricsStorage, rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)

	io.WriteString(rw, "<html><body>")
	io.WriteString(rw, "<h1>Metrics</h1>")

	io.WriteString(rw, "<h2>Gauges</h2><ul>")
	for name, value := range storage.GetAllGauges() {
		io.WriteString(rw, fmt.Sprintf("<li>%s: %v</li>", name, value))
	}
	io.WriteString(rw, "</ul>")

	io.WriteString(rw, "<h2>Counters</h2><ul>")
	for name, value := range storage.GetAllCounters() {
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

func (s *MemStorage) GetAllGauges() map[string]float64 {
	return s.gauges
}

func (s *MemStorage) GetAllCounters() map[string]int64 {
	return s.counters
}
