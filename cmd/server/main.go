package main

import (
	"io"
	"net/http"
	"strconv"

	"go-yandex-practicum/internal/model"

	"github.com/go-chi/chi/v5"
)

func main() {

	r := ConfigServerRouter()

	// r передаётся как http.Handler
	http.ListenAndServe(":8080", r)

}

func ConfigServerRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		io.WriteString(rw, "Hello, world!")
	})

	r.Route("/update", func(r chi.Router) {
		r.Route("/{metric-type}", func(r chi.Router) {
			r.Route("/{metric-name}", func(r chi.Router) {
				r.Post("/{metric-value}", metricHandler)
			})
		})
	})

	return r
}

func metricHandler(rw http.ResponseWriter, r *http.Request) {
	// тут работа с метрикой
	metricType := chi.URLParam(r, "metric-type")
	metricName := chi.URLParam(r, "metric-name")
	metricValue := chi.URLParam(r, "metric-value")

	switch metricType {
	case models.Counter:
		_, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(rw, "invalid counter value", http.StatusBadRequest)
			return
		}

	case models.Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, "invalid gauge value", http.StatusBadRequest)
			return
		}
		handleGauge(metricName, val)

	default:
		http.Error(rw, "unknown metric type", http.StatusBadRequest)
		return
	}
}

func handleGauge(metricName string, metricValue float64) {
	// здесь логика обработки gauge метрики
}
