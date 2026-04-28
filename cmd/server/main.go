package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-yandex-practicum/internal/config"
	"go-yandex-practicum/internal/middleware"
	"go-yandex-practicum/internal/model"
	"go-yandex-practicum/internal/service"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var AppConfig config.ServerConfig

func main() {
	parseFlags()

	if AppConfig.Restore {
		err := service.LoadMetricsFromFile(AppConfig.FileStorePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	if AppConfig.StoreInterval > 0 {
		go service.StoreMetrics(AppConfig.StoreInterval, AppConfig.FileStorePath)
	}

	r := ConfigServerRouter()

	_, port, err := net.SplitHostPort(AppConfig.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}

	addr := ":" + port

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func parseFlags() {
	flag.StringVar(&AppConfig.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&AppConfig.StoreInterval, "i", 300, "interval in seconds between metrics store")
	flag.StringVar(&AppConfig.FileStorePath, "f", "./metric-data", "path to store data")
	flag.BoolVar(&AppConfig.Restore, "r", false, "restore metric data")

	flag.Parse()

	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		parsedStoreInterval, err := strconv.Atoi(storeInterval)

		if err == nil {
			AppConfig.StoreInterval = parsedStoreInterval
		}
	}

	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		AppConfig.FileStorePath = filePath
	}

	if restore := os.Getenv("RESTORE"); restore != "" {
		parsedRestore, err := strconv.ParseBool(restore)

		if err == nil {
			AppConfig.Restore = parsedRestore
		}
	}

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		AppConfig.ServerAddress = envRunAddr
	}
}

func ConfigServerRouter() http.Handler {

	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.GzipMiddleware)

	r.Get("/", getMetricsListHandler)

	r.Route("/value", func(r chi.Router) {
		r.Post("/", getMetricValueJSONHandler)
		r.Route("/{metric-type}", func(r chi.Router) {
			r.Route("/{metric-name}", func(r chi.Router) {
				r.Get("/", getMetricValueHandler)
			})
		})
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/", metricJSONHandler)
		r.Route("/{metric-type}", func(r chi.Router) {
			r.Route("/{metric-name}", func(r chi.Router) {
				r.Post("/{metric-value}", metricHandler)
			})
		})
	})

	return r
}

func metricJSONHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(rw, "invalid Content-Type", http.StatusBadRequest)
		return
	}

	var m models.Metrics

	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(rw, "invalid JSON", http.StatusBadRequest)
		return
	}

	if m.ID == "" {
		http.Error(rw, "metric name required", http.StatusBadRequest)
		return
	}

	switch m.MType {
	case models.Counter:
		if m.Delta == nil {
			http.Error(rw, "delta required", http.StatusBadRequest)
			return
		}
		val := service.AddCounter(m.ID, *m.Delta)

		resp := models.Metrics{
			ID:    m.ID,
			MType: models.Counter,
			Delta: &val,
		}
		rw.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(rw).Encode(resp); err != nil {
			http.Error(rw, "encode response error", http.StatusInternalServerError)
			return
		}

	case models.Gauge:
		if m.Value == nil {
			http.Error(rw, "value required", http.StatusBadRequest)
			return
		}
		service.SetGauge(m.ID, *m.Value)

		resp := models.Metrics{
			ID:    m.ID,
			MType: models.Gauge,
			Value: m.Value,
		}

		rw.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(rw).Encode(resp); err != nil {
			http.Error(rw, "encode response error", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(rw, "unknown metric type", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func metricHandler(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric-type")
	metricName := chi.URLParam(r, "metric-name")
	metricValue := chi.URLParam(r, "metric-value")

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
		service.AddCounter(metricName, val)

	case models.Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, "invalid gauge value", http.StatusBadRequest)
			return
		}
		service.SetGauge(metricName, val)

	default:
		http.Error(rw, "unknown metric type", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func getMetricValueHandler(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric-type")
	metricName := chi.URLParam(r, "metric-name")

	switch metricType {
	case models.Counter:
		value, ok := service.GetCounter(metricName)
		if !ok {
			http.Error(rw, "unknown metric name", http.StatusNotFound)
			return
		}

		writeMetricValueResponse(rw, strconv.FormatInt(value, 10))
	case models.Gauge:
		value, ok := service.GetGauge(metricName)
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

func errorResponse(rw http.ResponseWriter, status int, msg string) {
	rw.Header().Set("Content-Type", "application/json")
	http.Error(rw, msg, status)
}

func getMetricValueJSONHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Header.Get("Content-Type") != "application/json" {
		errorResponse(rw, http.StatusNotFound, "invalid Content-Type")
		return
	}

	var req models.Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponse(rw, http.StatusNotFound, "invalid JSON")
		return
	}

	if req.ID == "" {
		errorResponse(rw, http.StatusNotFound, "metric name required")
		return
	}

	var resp models.Metrics

	switch req.MType {
	case models.Counter:
		value, ok := service.GetCounter(req.ID)
		if !ok {
			errorResponse(rw, http.StatusNotFound, "unknown metric name")
			return
		}

		resp = models.Metrics{
			ID:    req.ID,
			MType: req.MType,
			Delta: &value,
		}
	case models.Gauge:
		value, ok := service.GetGauge(req.ID)
		if !ok {
			errorResponse(rw, http.StatusNotFound, "unknown metric name")
			return
		}

		resp = models.Metrics{
			ID:    req.ID,
			MType: req.MType,
			Value: &value,
		}
	default:
		errorResponse(rw, http.StatusNotFound, "unknown metric type")
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(rw).Encode(resp); err != nil {
		errorResponse(rw, http.StatusInternalServerError, "encode response error")
		return
	}
}

func writeMetricValueResponse(rw http.ResponseWriter, metricValue string) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte(metricValue))
}

func getMetricsListHandler(rw http.ResponseWriter, r *http.Request) {
	buildMetricsListResponse(rw)
}

func buildMetricsListResponse(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)

	io.WriteString(rw, "<html><body>")
	io.WriteString(rw, "<h1>Metrics</h1>")

	io.WriteString(rw, "<h2>Gauges</h2><ul>")
	for name, value := range service.GetAllGauges() {
		io.WriteString(rw, fmt.Sprintf("<li>%s: %v</li>", name, value))
	}
	io.WriteString(rw, "</ul>")

	io.WriteString(rw, "<h2>Counters</h2><ul>")
	for name, value := range service.GetAllCounters() {
		io.WriteString(rw, fmt.Sprintf("<li>%s: %d</li>", name, value))
	}
	io.WriteString(rw, "</ul>")

	io.WriteString(rw, "</body></html>")
}
