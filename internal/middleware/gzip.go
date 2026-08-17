package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
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

func WithGzip(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			defer gzipReader.Close()

			r.Body = gzipReader
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzipWriter := &gzipResponseWriter{
				ResponseWriter: w,
			}

			h.ServeHTTP(gzipWriter, r)

			if gzipWriter.gzipWriter != nil {
				_ = gzipWriter.gzipWriter.Close()
			}

			return
		}

		h.ServeHTTP(w, r)
	})
}
