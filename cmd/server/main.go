package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-yandex-practicum/internal/config"
	"go-yandex-practicum/internal/middleware"
	"go-yandex-practicum/internal/model"
	"go-yandex-practicum/internal/repository"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	parseFlags()

	if AppConfig.Restore {
		err := loadMetricsFromFile(AppConfig.FileStorePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	if AppConfig.StoreInterval > 0 {
		go storeMetrics(AppConfig.StoreInterval, AppConfig.FileStorePath)
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

func storeMetrics(storeInterval int, filePath string) {
	ticker := time.NewTicker(time.Duration(storeInterval) * time.Second)

	defer ticker.Stop()
	for range ticker.C {
		if err := saveMetricsToFile(filePath); err != nil {
			log.Fatal(err)
		}
	}
}

func saveMetricsToFile(filePath string) error {

	metrics := make([]models.Metrics, 0)
	for name, value := range storage.GetAllGauges() {
		v := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &v,
		})
	}
	for name, value := range storage.GetAllCounters() {
		v := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &v,
		})
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(metrics)

}

func loadMetricsFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var metrics []models.Metrics
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		return err
	}

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				storage.SetGauge(metric.ID, *metric.Value)
			}
		case models.Counter:
			if metric.Delta != nil {
				storage.AddCounter(metric.ID, *metric.Delta)
			}
		}
	}

	return nil
}

var AppConfig config.ServerConfig

var storage repository.MetricsStorage = repository.NewMemStorage()

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
		val := storage.AddCounter(m.ID, *m.Delta)

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
		storage.SetGauge(m.ID, *m.Value)

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
	metricType := chi.URLParam(r, "metric-type")
	metricName := chi.URLParam(r, "metric-name")

	switch metricType {
	case models.Counter:
		value, ok := storage.GetCounter(metricName)
		if !ok {
			http.Error(rw, "unknown metric name", http.StatusNotFound)
			return
		}

		writeMetricValueResponse(rw, strconv.FormatInt(value, 10))
	case models.Gauge:
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
		value, ok := storage.GetCounter(req.ID)
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
		value, ok := storage.GetGauge(req.ID)
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

func writeMetricJSONValueResponse(rw http.ResponseWriter, metricType string, metricName string, metricValue float64) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	var resp = models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case models.Counter:
		v := int64(metricValue)
		resp.Delta = &v
	case models.Gauge:
		v := metricValue
		resp.Value = &v
	}

	err := json.NewEncoder(rw).Encode(resp)
	if err != nil {
		return
	}
}

func writeMetricValueResponse(rw http.ResponseWriter, metricValue string) {
	rw.Header().Set("Content-Type", "application/json")
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
