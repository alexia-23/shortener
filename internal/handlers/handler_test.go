package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	repository := &mockService{
		sourceLinks: make(map[string]string),
	}

	router := NewRouter(repository, "http://localhost:8080")

	require.NotNil(t, router)
}

func TestRouter_Routes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		links          map[string]string
		wantStatusCode int
	}{
		{
			name:           "create short link route",
			method:         http.MethodPost,
			path:           "/",
			body:           "https://example.com",
			links:          make(map[string]string),
			wantStatusCode: http.StatusCreated,
		},
		{
			name:   "get source link route",
			method: http.MethodGet,
			path:   "/MQ",
			links: map[string]string{
				"MQ": "https://example.com",
			},
			wantStatusCode: http.StatusTemporaryRedirect,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &mockService{
				createID:    "MQ",
				sourceLinks: test.links,
			}

			router := NewRouter(repository, "http://localhost:8080")

			request := httptest.NewRequest(
				test.method,
				test.path,
				strings.NewReader(test.body),
			)
			request.Header.Set("Content-Type", "text/plain")

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			assert.Equal(
				t,
				test.wantStatusCode,
				recorder.Code,
			)
		})
	}
}

func TestNewHandler_PanicsOnNilRepository(t *testing.T) {
	require.PanicsWithValue(
		t,
		"handlers: nil service",
		func() {
			NewHandler(nil, "http://localhost:8080")
		},
	)
}

func TestRouter_InvalidRequest(t *testing.T) {
	repository := &mockService{
		sourceLinks: make(map[string]string),
	}

	router := NewRouter(repository, "http://localhost:8080")

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "GET root",
			method: http.MethodGet,
			path:   "/",
		},
		{
			name:   "PUT root",
			method: http.MethodPut,
			path:   "/",
		},
		{
			name:   "POST with id",
			method: http.MethodPost,
			path:   "/MQ",
		},
		{
			name:   "unknown nested path",
			method: http.MethodGet,
			path:   "/one/two",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				test.method,
				test.path,
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			assert.Equal(
				t,
				http.StatusBadRequest,
				recorder.Code,
			)
		})
	}
}
