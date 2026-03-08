package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func setupRouter() http.Handler {
	r := chi.NewRouter()

	r.Route("/update", func(r chi.Router) {
		r.Route("/{metric-type}", func(r chi.Router) {
			r.Route("/{metric-name}", func(r chi.Router) {
				r.Post("/{metric-value}", metricHandler)
			})
		})
	})

	return r
}

func TestRootHandler(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("wrong status: want %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "Hello, world!" {
		t.Fatalf("wrong body: want %q, got %q", "Hello, world!", rec.Body.String())
	}
}

func TestMetricHandler_GaugeValid(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/123.45", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusBadRequest {
		t.Fatalf("did not expect bad request, got %d", rec.Code)
	}
}

func TestMetricHandler_GaugeInvalid(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/not-a-float", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("wrong status: want %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestMetricHandler_CounterValid(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/42", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusBadRequest {
		t.Fatalf("did not expect bad request, got %d", rec.Code)
	}
}

func TestMetricHandler_CounterInvalid(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/12.34", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("wrong status: want %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestMetricHandler_UnknownMetricType(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/update/unknown/AnyMetric/123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("wrong status: want %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestMetricHandler_WithoutMetricName_ReturnsNotFound(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/update/gauge//123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("wrong status: want %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestMetricHandler_WrongMethod_ReturnsMethodNotAllowed(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/update/gauge/Alloc/123.45", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("wrong status: want %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}
