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
	service := NewMockShortLinkService(t)

	router := NewRouter(service, "http://localhost:8080")

	require.NotNil(t, router)
}

func TestRouter_Routes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		setupMock      func(service *MockShortLinkService)
		wantStatusCode int
	}{
		{
			name:   "create short link route",
			method: http.MethodPost,
			path:   "/",
			body:   "https://example.com",
			setupMock: func(service *MockShortLinkService) {
				service.EXPECT().
					CreateShortLink("https://example.com").
					Return("MQ", nil).
					Once()
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:   "get source link route",
			method: http.MethodGet,
			path:   "/MQ",
			setupMock: func(service *MockShortLinkService) {
				service.EXPECT().
					GetSourceLink("MQ").
					Return("https://example.com", true).
					Once()
			},
			wantStatusCode: http.StatusTemporaryRedirect,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewMockShortLinkService(t)

			test.setupMock(service)

			router := NewRouter(service, "http://localhost:8080")

			request := httptest.NewRequest(
				test.method,
				test.path,
				strings.NewReader(test.body),
			)

			if test.method == http.MethodPost {
				request.Header.Set("Content-Type", "text/plain")
			}

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

func TestNewHandler_PanicsOnNilService(t *testing.T) {
	require.PanicsWithValue(
		t,
		"handlers: nil service",
		func() {
			NewHandler(nil, "http://localhost:8080")
		},
	)
}

func TestRouter_InvalidRequest(t *testing.T) {
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
			service := NewMockShortLinkService(t)

			router := NewRouter(service, "http://localhost:8080")

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
