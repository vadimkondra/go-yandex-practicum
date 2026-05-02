package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"go-yandex-practicum/internal/hash"
)

type hashResponseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *hashResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *hashResponseWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func HashMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" || key == "none" {
				next.ServeHTTP(w, r)
				return
			}

			if shouldCheckRequestHash(r) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "read body error", http.StatusBadRequest)
					return
				}
				defer r.Body.Close()

				requestHash := r.Header.Get(hash.HeaderName)
				if requestHash != "" && !hash.Check(body, key, requestHash) {
					http.Error(w, "invalid hash", http.StatusBadRequest)
					return
				}

				r.Body = io.NopCloser(bytes.NewReader(body))
			}

			hw := &hashResponseWriter{
				ResponseWriter: w,
				body:           bytes.NewBuffer(nil),
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(hw, r)

			responseBody := hw.body.Bytes()
			w.Header().Set(hash.HeaderName, hash.Calculate(responseBody, key))
			w.WriteHeader(hw.statusCode)
			_, _ = w.Write(responseBody)
		})
	}
}

func shouldCheckRequestHash(r *http.Request) bool {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		return false
	}

	return strings.HasPrefix(r.URL.Path, "/update/") ||
		strings.HasPrefix(r.URL.Path, "/updates/")
}
