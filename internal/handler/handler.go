package handler

import (
	"encoding/json"
	"fmt"
	"go-yandex-practicum/internal/middleware"
	"go-yandex-practicum/internal/model"
	"go-yandex-practicum/internal/service"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func ConfigServerRouter() http.Handler {

	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.GzipMiddleware)

	r.Get("/", GetMetricsListHandler)

	r.Get("/ping", PingHandler)

	r.Route("/value", func(r chi.Router) {
		r.Post("/", GetMetricValueJSONHandler)
		r.Route("/{metric-type}", func(r chi.Router) {
			r.Route("/{metric-name}", func(r chi.Router) {
				r.Get("/", GetMetricValueHandler)
			})
		})
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/", MetricJSONHandler)
		r.Route("/{metric-type}", func(r chi.Router) {
			r.Route("/{metric-name}", func(r chi.Router) {
				r.Post("/{metric-value}", MetricHandler)
			})
		})
	})

	r.Route("/updates", MetricsHandler)

	return r
}

func PingHandler(rw http.ResponseWriter, r *http.Request) {

	result, err := service.Ping()
	if err != nil || !result {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		rw.WriteHeader(http.StatusOK)
	}
}

func MetricJSONHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(rw, "invalid Content-Type", http.StatusBadRequest)
		return
	}

	var m model.Metrics

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
	case model.Counter:
		if m.Delta == nil {
			http.Error(rw, "delta required", http.StatusBadRequest)
			return
		}
		val, _ := service.AddCounter(m.ID, *m.Delta)

		resp := model.Metrics{
			ID:    m.ID,
			MType: model.Counter,
			Delta: &val,
		}
		rw.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(rw).Encode(resp); err != nil {
			http.Error(rw, "encode response error", http.StatusInternalServerError)
			return
		}

	case model.Gauge:
		if m.Value == nil {
			http.Error(rw, "value required", http.StatusBadRequest)
			return
		}
		err := service.SetGauge(m.ID, *m.Value)
		if err != nil {
			return
		}

		resp := model.Metrics{
			ID:    m.ID,
			MType: model.Gauge,
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

func MetricHandler(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric-type")
	metricName := chi.URLParam(r, "metric-name")
	metricValue := chi.URLParam(r, "metric-value")

	if metricName == "" {
		http.Error(rw, "metric name required", http.StatusNotFound)
		return
	}

	switch metricType {
	case model.Counter:
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(rw, "invalid counter value", http.StatusBadRequest)
			return
		}
		service.AddCounter(metricName, val)

	case model.Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, "invalid gauge value", http.StatusBadRequest)
			return
		}

		err = service.SetGauge(metricName, val)
		if err != nil {
			return
		}

	default:
		http.Error(rw, "unknown metric type", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func GetMetricValueHandler(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric-type")
	metricName := chi.URLParam(r, "metric-name")

	switch metricType {
	case model.Counter:
		value, ok, _ := service.GetCounter(metricName)
		if !ok {
			http.Error(rw, "unknown metric name", http.StatusNotFound)
			return
		}

		writeMetricValueResponse(rw, strconv.FormatInt(value, 10))
	case model.Gauge:
		value, ok, _ := service.GetGauge(metricName)
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

func GetMetricValueJSONHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Header.Get("Content-Type") != "application/json" {
		errorResponse(rw, http.StatusNotFound, "invalid Content-Type")
		return
	}

	var req model.Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponse(rw, http.StatusNotFound, "invalid JSON")
		return
	}

	if req.ID == "" {
		errorResponse(rw, http.StatusNotFound, "metric name required")
		return
	}

	var resp model.Metrics

	switch req.MType {
	case model.Counter:
		value, ok, _ := service.GetCounter(req.ID)
		if !ok {
			errorResponse(rw, http.StatusNotFound, "unknown metric name")
			return
		}

		resp = model.Metrics{
			ID:    req.ID,
			MType: req.MType,
			Delta: &value,
		}
	case model.Gauge:
		value, ok, _ := service.GetGauge(req.ID)
		if !ok {
			errorResponse(rw, http.StatusNotFound, "unknown metric name")
			return
		}

		resp = model.Metrics{
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

func GetMetricsListHandler(rw http.ResponseWriter, r *http.Request) {
	buildMetricsListResponse(rw)
}

func MetricsHandler(r chi.Router) {

}

func buildMetricsListResponse(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)

	io.WriteString(rw, "<html><body>")
	io.WriteString(rw, "<h1>Metrics</h1>")

	io.WriteString(rw, "<h2>Gauges</h2><ul>")

	gauges, _ := service.GetAllGauges()
	for name, value := range gauges {
		io.WriteString(rw, fmt.Sprintf("<li>%s: %v</li>", name, value))
	}
	io.WriteString(rw, "</ul>")

	io.WriteString(rw, "<h2>Counters</h2><ul>")

	counters, _ := service.GetAllCounters()
	for name, value := range counters {
		io.WriteString(rw, fmt.Sprintf("<li>%s: %d</li>", name, value))
	}
	io.WriteString(rw, "</ul>")

	io.WriteString(rw, "</body></html>")
}

func errorResponse(rw http.ResponseWriter, status int, msg string) {
	rw.Header().Set("Content-Type", "application/json")
	http.Error(rw, msg, status)
}

func writeMetricValueResponse(rw http.ResponseWriter, metricValue string) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte(metricValue))
}
