package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_handleCreateShortLinkJSON_Success(t *testing.T) {
	service := NewMockShortLinkService(t)

	service.EXPECT().
		CreateShortLink(
			mock.Anything,
			"https://example.com",
		).
		Return("MQ", nil).
		Once()

	router := NewRouter(service, "http://localhost:8080")

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(`{"url":"https://example.com"}`),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	assert.Equal(
		t,
		"application/json",
		recorder.Header().Get("Content-Type"),
	)

	var response createShortLinkResponse

	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(
		t,
		"http://localhost:8080/MQ",
		response.Result,
	)
}

func TestHandler_handleCreateShortLinkJSON_ServiceError(t *testing.T) {
	service := NewMockShortLinkService(t)

	service.EXPECT().
		CreateShortLink(
			mock.Anything,
			"https://example.com",
		).
		Return("", errors.New("service error")).
		Once()

	router := NewRouter(service, "http://localhost:8080")

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(`{"url":"https://example.com"}`),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestHandler_handleCreateShortLinkJSON_URLWithSpaces(t *testing.T) {
	service := NewMockShortLinkService(t)

	service.EXPECT().
		CreateShortLink(
			mock.Anything,
			"https://example.com",
		).
		Return("MQ", nil).
		Once()

	router := NewRouter(service, "http://localhost:8080")

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(`{"url":"   https://example.com   "}`),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)
}

func TestHandler_handleCreateShortLinkJSON_BaseURLWithTrailingSlash(
	t *testing.T,
) {
	service := NewMockShortLinkService(t)

	service.EXPECT().
		CreateShortLink(
			mock.Anything,
			"https://example.com",
		).
		Return("MQ", nil).
		Once()

	router := NewRouter(service, "http://localhost:8080/")

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(`{"url":"https://example.com"}`),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response createShortLinkResponse

	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(
		t,
		"http://localhost:8080/MQ",
		response.Result,
	)
}

func TestHandler_handleCreateShortLinkJSON_EmptyBody(t *testing.T) {
	service := NewMockShortLinkService(t)

	router := NewRouter(service, "http://localhost:8080")

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		nil,
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandler_handleCreateShortLinkJSON_MethodNotAllowed(t *testing.T) {
	service := NewMockShortLinkService(t)

	router := NewRouter(service, "http://localhost:8080")

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/shorten",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandler_handleCreateShortLinkJSON_UppercaseURLKey(t *testing.T) {
	service := NewMockShortLinkService(t)

	service.EXPECT().
		CreateShortLink(
			mock.Anything,
			"https://example.com",
		).
		Return("MQ", nil).
		Once()

	router := NewRouter(service, "http://localhost:8080")

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(`{"URL":"https://example.com"}`),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)
}

func TestHandler_handleCreateShortLinkJSON_UnknownField(t *testing.T) {
	service := NewMockShortLinkService(t)

	service.EXPECT().
		CreateShortLink(
			mock.Anything,
			"https://example.com",
		).
		Return("MQ", nil).
		Once()

	router := NewRouter(service, "http://localhost:8080")

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(`{
			"url":"https://example.com",
			"foo":"bar"
		}`),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)
}

func TestHandler_handleCreateShortLinkJSON_BadRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "invalid JSON",
			body: `{"url":`,
		},
		{
			name: "empty URL",
			body: `{"url":""}`,
		},
		{
			name: "invalid URL",
			body: `{"url":"hello"}`,
		},
		{
			name: "missing URL",
			body: `{}`,
		},
		{
			name: "URL with only spaces",
			body: `{"url":"     "}`,
		},
		{
			name: "invalid URL type",
			body: `{"url":123}`,
		},
		{
			name: "null URL",
			body: `{"url":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewMockShortLinkService(t)

			router := NewRouter(service, "http://localhost:8080")

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/shorten",
				strings.NewReader(tt.body),
			)

			request.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			assert.Equal(t, http.StatusBadRequest, recorder.Code)
		})
	}
}
