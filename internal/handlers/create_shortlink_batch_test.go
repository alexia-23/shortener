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
	"github.com/stretchr/testify/require"
)

func TestHandler_handleCreateShortLinksBatch_Success(t *testing.T) {
	service := NewMockShortLinkService(t)

	service.EXPECT().
		CreateShortLinksBatch(
			mock.Anything,
			[]string{
				"https://example.com/first",
				"https://example.com/second",
			},
		).
		Return([]string{"first-id", "second-id"}, nil).
		Once()

	router := NewRouter(service, "http://localhost:8080/")
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten/batch",
		strings.NewReader(`[
			{
				"correlation_id":"first-correlation",
				"original_url":"https://example.com/first"
			},
			{
				"correlation_id":"second-correlation",
				"original_url":"https://example.com/second"
			}
		]`),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(
		t,
		"application/json",
		recorder.Header().Get("Content-Type"),
	)

	var response []createShortLinksBatchResponseItem
	require.NoError(
		t,
		json.NewDecoder(recorder.Body).Decode(&response),
	)

	assert.Equal(
		t,
		[]createShortLinksBatchResponseItem{
			{
				CorrelationID: "first-correlation",
				ShortURL:      "http://localhost:8080/first-id",
			},
			{
				CorrelationID: "second-correlation",
				ShortURL:      "http://localhost:8080/second-id",
			},
		},
		response,
	)
}

func TestHandler_handleCreateShortLinksBatch_BadRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "invalid JSON", body: `[{`},
		{name: "empty batch", body: `[]`},
		{name: "null batch", body: `null`},
		{
			name: "empty original URL",
			body: `[{
				"correlation_id":"first",
				"original_url":""
			}]`,
		},
		{
			name: "invalid original URL",
			body: `[{
				"correlation_id":"first",
				"original_url":"not-a-url"
			}]`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewMockShortLinkService(t)
			router := NewRouter(
				service,
				"http://localhost:8080",
			)

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/shorten/batch",
				strings.NewReader(test.body),
			)
			request.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			assert.Equal(t, http.StatusBadRequest, recorder.Code)
		})
	}
}

func TestHandler_handleCreateShortLinksBatch_ServiceError(t *testing.T) {
	service := NewMockShortLinkService(t)

	service.EXPECT().
		CreateShortLinksBatch(
			mock.Anything,
			[]string{"https://example.com"},
		).
		Return(nil, errors.New("service error")).
		Once()

	router := NewRouter(service, "http://localhost:8080")
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten/batch",
		strings.NewReader(`[{
			"correlation_id":"first",
			"original_url":"https://example.com"
		}]`),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	assert.Equal(
		t,
		http.StatusInternalServerError,
		recorder.Code,
	)
}
