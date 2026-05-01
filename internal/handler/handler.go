package handler

import (
	"encoding/json"
	"fmt"
	"go-yandex-practicum/internal/config"
	"go-yandex-practicum/internal/middleware"
	"go-yandex-practicum/internal/model"
	"go-yandex-practicum/internal/service"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ServerHandler struct {
	service *service.MetricsService
	cfg     config.ServerConfig
}

func NewServerHandler(service *service.MetricsService, cfg config.ServerConfig) *ServerHandler {
	return &ServerHandler{service: service, cfg: cfg}
}

func ConfigServerRouter(service *service.MetricsService, cfg config.ServerConfig) http.Handler {
	h := NewServerHandler(service, cfg)

	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.HashMiddleware(cfg.Key))
	r.Use(middleware.GzipMiddleware)

	r.Get("/", h.GetMetricsListHandler)

	r.Get("/ping", h.PingHandler)

	r.Route("/value", func(r chi.Router) {
		r.Post("/", h.GetMetricValueJSONHandler)
		r.Route("/{metric-type}", func(r chi.Router) {
			r.Route("/{metric-name}", func(r chi.Router) {
				r.Get("/", h.GetMetricValueHandler)
			})
		})
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.MetricJSONHandler)
		r.Route("/{metric-type}", func(r chi.Router) {
			r.Route("/{metric-name}", func(r chi.Router) {
				r.Post("/{metric-value}", h.MetricHandler)
			})
		})
	})

	r.Post("/updates/", h.MetricsHandler)

	return r
}

func (h *ServerHandler) PingHandler(rw http.ResponseWriter, r *http.Request) {

	result, err := h.service.Ping()
	if err != nil || !result {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		rw.WriteHeader(http.StatusOK)
	}
}

func (h *ServerHandler) MetricJSONHandler(rw http.ResponseWriter, r *http.Request) {
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
		val, _ := h.service.AddCounter(m.ID, *m.Delta)

		resp := model.Metrics{
			ID:    m.ID,
			MType: model.Counter,
			Delta: &val,
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(rw).Encode(resp); err != nil {
			http.Error(rw, "encode response error", http.StatusInternalServerError)
			return
		}
		return

	case model.Gauge:
		if m.Value == nil {
			http.Error(rw, "value required", http.StatusBadRequest)
			return
		}
		err := h.service.SetGauge(m.ID, *m.Value)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
		}

		resp := model.Metrics{
			ID:    m.ID,
			MType: model.Gauge,
			Value: m.Value,
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(rw).Encode(resp); err != nil {
			http.Error(rw, "encode response error", http.StatusInternalServerError)
			return
		}
		return
	default:
		http.Error(rw, "unknown metric type", http.StatusBadRequest)
		return
	}
}

func (h *ServerHandler) MetricHandler(rw http.ResponseWriter, r *http.Request) {
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
		h.service.AddCounter(metricName, val)

	case model.Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, "invalid gauge value", http.StatusBadRequest)
			return
		}

		err = h.service.SetGauge(metricName, val)
		if err != nil {
			return
		}

	default:
		http.Error(rw, "unknown metric type", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *ServerHandler) GetMetricValueHandler(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric-type")
	metricName := chi.URLParam(r, "metric-name")

	switch metricType {
	case model.Counter:
		value, ok, _ := h.service.GetCounter(metricName)
		if !ok {
			http.Error(rw, "unknown metric name", http.StatusNotFound)
			return
		}

		writeMetricValueResponse(rw, strconv.FormatInt(value, 10))
	case model.Gauge:
		value, ok, _ := h.service.GetGauge(metricName)
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

func (h *ServerHandler) GetMetricValueJSONHandler(rw http.ResponseWriter, r *http.Request) {
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
		value, ok, _ := h.service.GetCounter(req.ID)
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
		value, ok, _ := h.service.GetGauge(req.ID)
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

func (h *ServerHandler) GetMetricsListHandler(rw http.ResponseWriter, r *http.Request) {
	h.buildMetricsListResponse(rw)
}

func (h *ServerHandler) MetricsHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(rw, "invalid Content-Type", http.StatusBadRequest)
		return
	}

	var metrics []model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(rw, "invalid JSON", http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		rw.WriteHeader(http.StatusOK)
		return
	}

	updatedMetrics, err := h.service.UpdateMetricsBatch(metrics)
	if err != nil {
		http.Error(rw, "update metrics batch error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(rw).Encode(updatedMetrics); err != nil {
		http.Error(rw, "encode response error", http.StatusInternalServerError)
		return
	}
}

func (h *ServerHandler) buildMetricsListResponse(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)

	io.WriteString(rw, "<html><body>")
	io.WriteString(rw, "<h1>Metrics</h1>")

	io.WriteString(rw, "<h2>Gauges</h2><ul>")

	gauges, _ := h.service.GetAllGauges()
	for name, value := range gauges {
		io.WriteString(rw, fmt.Sprintf("<li>%s: %v</li>", name, value))
	}
	io.WriteString(rw, "</ul>")

	io.WriteString(rw, "<h2>Counters</h2><ul>")

	counters, _ := h.service.GetAllCounters()
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
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(metricValue))
}
