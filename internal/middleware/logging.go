package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func WithLogging(
	h http.Handler,
	sugar *zap.SugaredLogger,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData: &responseData{
				status: http.StatusOK,
			},
		}
		h.ServeHTTP(lrw, r)

		sugar.Infow(
			"request",
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", time.Since(start),
			"status", lrw.responseData.status,
			"size", lrw.responseData.size,
		)
	})
}
