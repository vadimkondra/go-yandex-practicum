package middleware

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

var sugar zap.SugaredLogger

func LoggingMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logger, err := zap.NewDevelopment()
		if err != nil {
			// вызываем панику, если ошибка
			panic(err)
		}
		defer logger.Sync()

		// делаем регистратор SugaredLogger
		sugar = *logger.Sugar()

		start := time.Now()

		lw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		handler.ServeHTTP(lw, r)

		duration := time.Since(start)

		sugar.Infow(
			"request completed",
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
			"status", lw.statusCode,
			"size", lw.size,
		)
	})
}
