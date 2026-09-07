package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	shortlinkservice "github.com/alexia-23/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("read error")
}

func TestHandler_handleCreateShortLink(t *testing.T) {
	tests := []struct {
		name            string
		body            func() io.Reader
		contentType     string
		wantStatusCode  int
		wantBody        string
		wantContentType string
		wantSavedURL    string
	}{
		{
			name: "successful request",
			body: func() io.Reader {
				return strings.NewReader("https://example.com")
			},
			contentType:     "text/plain",
			wantStatusCode:  http.StatusCreated,
			wantBody:        "http://localhost:8080/MQ",
			wantContentType: "text/plain",
			wantSavedURL:    "https://example.com",
		},
		{
			name: "empty body",
			body: func() io.Reader {
				return strings.NewReader("")
			},
			contentType:     "text/plain",
			wantStatusCode:  http.StatusBadRequest,
			wantBody:        "",
			wantContentType: "",
			wantSavedURL:    "",
		},
		{
			name: "invalid URL",
			body: func() io.Reader {
				return strings.NewReader("hello")
			},
			contentType:     "text/plain",
			wantStatusCode:  http.StatusBadRequest,
			wantBody:        "",
			wantContentType: "",
			wantSavedURL:    "",
		},
		{
			name: "URL with spaces",
			body: func() io.Reader {
				return strings.NewReader("  https://example.com  ")
			},
			contentType:     "text/plain",
			wantStatusCode:  http.StatusCreated,
			wantBody:        "http://localhost:8080/MQ",
			wantContentType: "text/plain",
			wantSavedURL:    "https://example.com",
		},
		{
			name: "text plain with charset",
			body: func() io.Reader {
				return strings.NewReader("https://example.com")
			},
			contentType:     "text/plain; charset=utf-8",
			wantStatusCode:  http.StatusCreated,
			wantBody:        "http://localhost:8080/MQ",
			wantContentType: "text/plain",
			wantSavedURL:    "https://example.com",
		},
		{
			name: "wrong content type",
			body: func() io.Reader {
				return strings.NewReader("https://example.com")
			},
			contentType:     "application/json",
			wantStatusCode:  http.StatusBadRequest,
			wantBody:        "",
			wantContentType: "",
			wantSavedURL:    "",
		},
		{
			name: "body reading error",
			body: func() io.Reader {
				return errorReader{}
			},
			contentType:     "text/plain",
			wantStatusCode:  http.StatusBadRequest,
			wantBody:        "",
			wantContentType: "",
			wantSavedURL:    "",
		},
		{
			name: "body is too large",
			body: func() io.Reader {
				return strings.NewReader(
					"https://example.com/" +
						strings.Repeat("a", int(maxRequestBodySize)),
				)
			},
			contentType:     "text/plain",
			wantStatusCode:  http.StatusBadRequest,
			wantBody:        "",
			wantContentType: "",
			wantSavedURL:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewMockShortLinkService(t)

			if test.wantSavedURL != "" {
				service.EXPECT().
					CreateShortLink(
						mock.Anything,
						test.wantSavedURL,
					).
					Return("MQ", nil).
					Once()
			}

			router := NewRouter(service, "http://localhost:8080")

			request := httptest.NewRequest(
				http.MethodPost,
				"/",
				test.body(),
			)

			if test.contentType != "" {
				request.Header.Set("Content-Type", test.contentType)
			}

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			responseBody, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			assert.Equal(
				t,
				test.wantStatusCode,
				response.StatusCode,
			)

			assert.Equal(
				t,
				test.wantBody,
				string(responseBody),
			)

			assert.Equal(
				t,
				test.wantContentType,
				response.Header.Get("Content-Type"),
			)
		})
	}
}

func TestHandler_handleCreateShortLink_OriginalURLExists(t *testing.T) {
	mockService := NewMockShortLinkService(t)

	mockService.EXPECT().
		CreateShortLink(
			mock.Anything,
			"https://example.com",
		).
		Return(
			"",
			fmt.Errorf(
				"save short link: %w",
				&shortlinkservice.OriginalURLExistsError{
					ID: "existing-id",
				},
			),
		).
		Once()

	router := NewRouter(mockService, "http://localhost:8080")

	request := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("https://example.com"),
	)
	request.Header.Set("Content-Type", "text/plain")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.Equal(
		t,
		"text/plain",
		recorder.Header().Get("Content-Type"),
	)
	assert.Equal(
		t,
		"http://localhost:8080/existing-id",
		recorder.Body.String(),
	)
}
