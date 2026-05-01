package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddlewareCompressesJSONResponse(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			t.Fatalf("write response error = %v", err)
		}
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected Content-Encoding gzip, got %q", response.Header.Get("Content-Encoding"))
	}

	gzReader, err := gzip.NewReader(response.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader() error = %v", err)
	}
	defer gzReader.Close()

	body, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("read gzip body error = %v", err)
	}

	if string(body) != `{"status":"ok"}` {
		t.Fatalf("expected body %q, got %q", `{"status":"ok"}`, string(body))
	}
}

func TestGzipMiddlewareCompressesHTMLResponse(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("<html><body>Hello</body></html>"))
		if err != nil {
			t.Fatalf("write response error = %v", err)
		}
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected Content-Encoding gzip, got %q", response.Header.Get("Content-Encoding"))
	}

	gzReader, err := gzip.NewReader(response.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader() error = %v", err)
	}
	defer gzReader.Close()

	body, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("read gzip body error = %v", err)
	}

	expected := "<html><body>Hello</body></html>"
	if string(body) != expected {
		t.Fatalf("expected body %q, got %q", expected, string(body))
	}
}

func TestGzipMiddlewareDoesNotCompressUnsupportedContentType(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("plain text"))
		if err != nil {
			t.Fatalf("write response error = %v", err)
		}
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.Header.Get("Content-Encoding") == "gzip" {
		t.Fatal("did not expect gzip Content-Encoding for text/plain")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body error = %v", err)
	}

	if string(body) != "plain text" {
		t.Fatalf("expected body %q, got %q", "plain text", string(body))
	}
}

func TestGzipMiddlewareDoesNotCompressWhenClientDoesNotSupportGzip(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			t.Fatalf("write response error = %v", err)
		}
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.Header.Get("Content-Encoding") == "gzip" {
		t.Fatal("did not expect gzip Content-Encoding without Accept-Encoding gzip")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body error = %v", err)
	}

	if string(body) != `{"status":"ok"}` {
		t.Fatalf("expected body %q, got %q", `{"status":"ok"}`, string(body))
	}
}

func TestGzipMiddlewareDecompressesGzipRequestBody(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body error = %v", err)
		}

		if string(body) != `{"metric":"Alloc"}` {
			t.Fatalf("expected request body %q, got %q", `{"metric":"Alloc"}`, string(body))
		}

		w.WriteHeader(http.StatusOK)
	}))

	var buf bytes.Buffer

	gzWriter := gzip.NewWriter(&buf)
	_, err := gzWriter.Write([]byte(`{"metric":"Alloc"}`))
	if err != nil {
		t.Fatalf("gzip write error = %v", err)
	}

	if err := gzWriter.Close(); err != nil {
		t.Fatalf("gzip close error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/", &buf)
	request.Header.Set("Content-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
}

func TestGzipMiddlewareReturnsBadRequestForInvalidGzipBody(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called for invalid gzip body")
	}))

	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not gzip"))
	request.Header.Set("Content-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}
}
