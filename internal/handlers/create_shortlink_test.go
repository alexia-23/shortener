package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			wantStatusCode:  http.StatusBadRequest,
			wantBody:        "",
			wantContentType: "",
			wantSavedURL:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &mockRepository{
				saveID: "MQ",
				links:  make(map[string]string),
			}

			handler := NewHandler(repository)

			request := httptest.NewRequest(
				http.MethodPost,
				"/",
				test.body(),
			)

			request.Header.Set("Content-Type", "text/plain")

			mux := http.NewServeMux()
			handler.RegisterRoutes(mux)

			recorder := httptest.NewRecorder()

			mux.ServeHTTP(recorder, request)

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

			assert.Equal(
				t,
				test.wantSavedURL,
				repository.savedURL,
			)
		})
	}
}
