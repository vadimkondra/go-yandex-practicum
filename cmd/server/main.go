package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func main() {

	r := chi.NewRouter()

	r.Route("/update", func(r chi.Router) {
		r.Route("/{metric-type}", func(r chi.Router) {
			r.Route("/{metric-name}", func(r chi.Router) {
				r.Post("/{metric-value}", metricHandler)
			})
		})
	})

	// r передаётся как http.Handler
	http.ListenAndServe(":8080", r)
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
	case "counter":
		_, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(rw, "invalid counter value", http.StatusBadRequest)
			return
		}

	case "gauge":
		_, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, "invalid gauge value", http.StatusBadRequest)
			return
		}

	default:
		http.Error(rw, "unknown metric type", http.StatusBadRequest)
		return
	}
}
