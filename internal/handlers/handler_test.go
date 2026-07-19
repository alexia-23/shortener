package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	repository := &mockRepository{
		links: make(map[string]string),
	}

	handler := NewHandler(repository)

	require.NotNil(t, handler)
}

func TestHandler_RegisterRoutes(t *testing.T) {
	repository := &mockRepository{
		links: make(map[string]string),
	}

	handler := NewHandler(repository)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	tests := []struct {
		name        string
		method      string
		path        string
		wantPattern string
	}{
		{
			name:        "create short link route",
			method:      http.MethodPost,
			path:        "/",
			wantPattern: "POST /{$}",
		},
		{
			name:        "get source link route",
			method:      http.MethodGet,
			path:        "/MQ",
			wantPattern: "GET /{id}",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				test.method,
				test.path,
				nil,
			)

			_, pattern := mux.Handler(request)

			assert.Equal(t, test.wantPattern, pattern)
		})
	}
}

func TestHandler_InvalidRequest(t *testing.T) {
	repository := &mockRepository{
		links: make(map[string]string),
	}

	handler := NewHandler(repository)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

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

			mux.ServeHTTP(recorder, request)

			assert.Equal(
				t,
				http.StatusBadRequest,
				recorder.Code,
			)
		})
	}
}
