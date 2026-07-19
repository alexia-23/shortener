package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_handleGetSourceLink(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		links          map[string]string
		wantStatusCode int
		wantLocation   string
	}{
		{
			name: "link exists",
			id:   "MQ",
			links: map[string]string{
				"MQ": "https://example.com",
			},
			wantStatusCode: http.StatusTemporaryRedirect,
			wantLocation:   "https://example.com",
		},
		{
			name:           "link does not exist",
			id:             "unknown",
			links:          make(map[string]string),
			wantStatusCode: http.StatusBadRequest,
			wantLocation:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &mockRepository{
				links: test.links,
			}

			handler := NewHandler(repository)

			request := httptest.NewRequest(
				http.MethodGet,
				"/"+test.id,
				nil,
			)

			request.SetPathValue("id", test.id)

			recorder := httptest.NewRecorder()

			handler.handleGetSourceLink(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			assert.Equal(
				t,
				test.wantStatusCode,
				response.StatusCode,
			)

			assert.Equal(
				t,
				test.wantLocation,
				response.Header.Get("Location"),
			)
		})
	}
}
