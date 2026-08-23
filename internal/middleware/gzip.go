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

type gzipMiddleware struct {
	sugar *zap.SugaredLogger
}

type gzipHandler struct {
	next  http.Handler
	sugar *zap.SugaredLogger
}

func WithGzip(
	sugar *zap.SugaredLogger,
) func(http.Handler) http.Handler {
	middleware := gzipMiddleware{
		sugar: sugar,
	}

	return middleware.wrap
}

func (middleware gzipMiddleware) wrap(
	next http.Handler,
) http.Handler {
	return &gzipHandler{
		next:  next,
		sugar: middleware.sugar,
	}
}

func (handler *gzipHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	gzipReader, err := gzipDecode(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if gzipReader != nil {
		defer handler.closeGzipReader(gzipReader)
	}

	gzipWriter := gzipEncode(w, r)
	if gzipWriter == nil {
		handler.next.ServeHTTP(w, r)
		return
	}

	defer handler.closeGzipWriter(gzipWriter)

	handler.next.ServeHTTP(gzipWriter, r)
}

func gzipDecode(
	r *http.Request,
) (*gzip.Reader, error) {
	if r.Header.Get("Content-Encoding") != "gzip" {
		return nil, nil
	}

	gzipReader, err := gzip.NewReader(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = gzipReader

	return gzipReader, nil
}

func gzipEncode(
	w http.ResponseWriter,
	r *http.Request,
) *gzipResponseWriter {
	if !strings.Contains(
		r.Header.Get("Accept-Encoding"),
		"gzip",
	) {
		return nil
	}

	return &gzipResponseWriter{
		ResponseWriter: w,
	}
}

func (handler *gzipHandler) closeGzipReader(
	gzipReader *gzip.Reader,
) {
	if err := gzipReader.Close(); err != nil {
		handler.sugar.Errorw(
			"failed to close gzip request reader",
			"error", err,
		)
	}
}

func (handler *gzipHandler) closeGzipWriter(
	responseWriter *gzipResponseWriter,
) {
	if responseWriter.gzipWriter == nil {
		return
	}

	if err := responseWriter.gzipWriter.Close(); err != nil {
		handler.sugar.Errorw(
			"failed to close gzip response writer",
			"error", err,
		)
	}
}

func (w *gzipResponseWriter) Write(
	data []byte,
) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if w.compress {
		return w.gzipWriter.Write(data)
	}

	return w.ResponseWriter.Write(data)
}

func (w *gzipResponseWriter) WriteHeader(
	statusCode int,
) {
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

func isCompressibleContentType(
	contentType string,
) bool {
	return strings.HasPrefix(
		contentType,
		"application/json",
	) || strings.HasPrefix(
		contentType,
		"text/html",
	)
}
