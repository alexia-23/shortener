package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithGzip_DecompressRequest(t *testing.T) {
	originalBody := []byte(`{"url":"https://example.com"}`)

	var compressedBody bytes.Buffer

	gzipWriter := gzip.NewWriter(&compressedBody)

	_, err := gzipWriter.Write(originalBody)
	assert.NoError(t, err)

	err = gzipWriter.Close()
	assert.NoError(t, err)

	var receivedBody []byte

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, err = io.ReadAll(r.Body)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	})

	gzipHandler := WithGzip(handler)

	request := httptest.NewRequest(
		http.MethodPost,
		"/",
		bytes.NewReader(compressedBody.Bytes()),
	)

	request.Header.Set("Content-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	gzipHandler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, originalBody, receivedBody)
}
func TestWithGzip_UncompressedRequest(t *testing.T) {
	originalBody := []byte(`{"url":"https://example.com"}`)

	var receivedBody []byte

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		receivedBody, err = io.ReadAll(r.Body)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	})

	gzipHandler := WithGzip(handler)

	request := httptest.NewRequest(
		http.MethodPost,
		"/",
		bytes.NewReader(originalBody),
	)

	recorder := httptest.NewRecorder()

	gzipHandler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, originalBody, receivedBody)
}
func TestWithGzip_InvalidCompressedRequest(t *testing.T) {
	handlerCalled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	gzipHandler := WithGzip(handler)

	request := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("this is not gzip"),
	)

	request.Header.Set("Content-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	gzipHandler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.False(t, handlerCalled)
}
func TestWithGzip_CompressJSONResponse(t *testing.T) {
	originalBody := []byte(`{"result":"http://localhost:8080/MQ"}`)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		_, err := w.Write(originalBody)
		assert.NoError(t, err)
	})

	gzipHandler := WithGzip(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	request.Header.Set("Accept-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	gzipHandler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(
		t,
		"gzip",
		recorder.Header().Get("Content-Encoding"),
	)

	gzipReader, err := gzip.NewReader(recorder.Body)
	assert.NoError(t, err)

	if err != nil {
		return
	}

	defer gzipReader.Close()

	body, err := io.ReadAll(gzipReader)
	assert.NoError(t, err)

	assert.Equal(t, originalBody, body)
}

func TestWithGzip_CompressHTMLResponse(t *testing.T) {
	originalBody := []byte(`<html><body>Hello</body></html>`)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write(originalBody)
		assert.NoError(t, err)
	})

	gzipHandler := WithGzip(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	request.Header.Set("Accept-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	gzipHandler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(
		t,
		"gzip",
		recorder.Header().Get("Content-Encoding"),
	)

	gzipReader, err := gzip.NewReader(recorder.Body)
	assert.NoError(t, err)

	if err != nil {
		return
	}

	defer gzipReader.Close()

	body, err := io.ReadAll(gzipReader)
	assert.NoError(t, err)

	assert.Equal(t, originalBody, body)
}
func TestWithGzip_DoesNotCompressTextPlain(t *testing.T) {
	originalBody := []byte("hello")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write(originalBody)
		assert.NoError(t, err)
	})

	gzipHandler := WithGzip(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	request.Header.Set("Accept-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	gzipHandler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	assert.Empty(
		t,
		recorder.Header().Get("Content-Encoding"),
	)

	assert.Equal(
		t,
		originalBody,
		recorder.Body.Bytes(),
	)
}
func TestWithGzip_DoesNotCompressWithoutAcceptEncoding(t *testing.T) {
	originalBody := []byte(`{"result":"http://localhost:8080/MQ"}`)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		_, err := w.Write(originalBody)
		assert.NoError(t, err)
	})

	gzipHandler := WithGzip(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	recorder := httptest.NewRecorder()

	gzipHandler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	assert.Empty(
		t,
		recorder.Header().Get("Content-Encoding"),
	)

	assert.Equal(
		t,
		originalBody,
		recorder.Body.Bytes(),
	)
}
func TestWithGzip_CompressJSONResponseWithCharset(t *testing.T) {
	originalBody := []byte(`{"result":"ok"}`)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write(originalBody)
		assert.NoError(t, err)
	})

	gzipHandler := WithGzip(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	request.Header.Set("Accept-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	gzipHandler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(
		t,
		"gzip",
		recorder.Header().Get("Content-Encoding"),
	)

	gzipReader, err := gzip.NewReader(recorder.Body)
	assert.NoError(t, err)

	if err != nil {
		return
	}

	defer gzipReader.Close()

	body, err := io.ReadAll(gzipReader)
	assert.NoError(t, err)

	assert.Equal(t, originalBody, body)
}
