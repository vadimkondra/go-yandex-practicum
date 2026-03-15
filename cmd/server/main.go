package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"go-yandex-practicum/internal/model"
	"go-yandex-practicum/internal/repository"
	"go-yandex-practicum/internal/config"

	"github.com/go-chi/chi/v5"
)

func main() {
	parseFlags()

	r := ConfigServerRouter()

	if err := http.ListenAndServe(AppConfig.ServerAddress, r); err != nil {
		panic(err)
	}
}

var AppConfig config.ServerConfig

const (
	metricTypeRouteName  = "metric-type"
	metricNameRouteName  = "metric-name"
	metricValueRouteName = "metric-value"
)

var storage repository.MetricsStorage = repository.NewMemStorage()


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
		storage.AddCounter(metricName, val)

	case models.Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, "invalid gauge value", http.StatusBadRequest)
			return
		}
		storage.SetGauge(metricName, val)

	default:
		http.Error(rw, "unknown metric type", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func getMetricValueHandler(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, metricTypeRouteName)
	metricName := chi.URLParam(r, metricNameRouteName)

	switch metricType {
	case "counter":
		value, ok := storage.GetCounter(metricName)
		if !ok {
			http.Error(rw, "unknown metric name", http.StatusNotFound)
			return
		}

		writeMetricValueResponse(rw, strconv.FormatInt(value, 10))
	case "gauge":
		value, ok := storage.GetGauge(metricName)
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

func buildMetricsListResponse(storage repository.MetricsStorage, rw http.ResponseWriter) {
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
