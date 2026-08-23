package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter  *gzip.Writer
	compress    bool
	wroteHeader bool
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if w.compress {
		return w.gzipWriter.Write(data)
	}

	return w.ResponseWriter.Write(data)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	contentType := w.Header().Get("Content-Type")

	if isCompressibleContentType(contentType) {
		w.compress = true
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")

		w.gzipWriter = gzip.NewWriter(w.ResponseWriter)
	}

	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func isCompressibleContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "text/html")
}

func WithGzip(
	sugar *zap.SugaredLogger,
) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Encoding") == "gzip" {
				gzipReader, err := gzip.NewReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				defer func() {
					if err := gzipReader.Close(); err != nil {
						sugar.Errorw(
							"failed to close gzip request reader",
							"error", err,
						)
					}
				}()

				r.Body = gzipReader
			}

			if strings.Contains(
				r.Header.Get("Accept-Encoding"),
				"gzip",
			) {
				gzipWriter := &gzipResponseWriter{
					ResponseWriter: w,
				}

				defer func() {
					if gzipWriter.gzipWriter == nil {
						return
					}

					if err := gzipWriter.gzipWriter.Close(); err != nil {
						sugar.Errorw(
							"failed to close gzip response writer",
							"error", err,
						)
					}
				}()

				h.ServeHTTP(gzipWriter, r)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
