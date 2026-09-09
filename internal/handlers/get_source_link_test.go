package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_handleGetSourceLink(t *testing.T) {
	tests := []struct {
		name            string
		id              string
		wantOriginalURL string
		wantFound       bool
		wantStatusCode  int
		wantLocation    string
	}{
		{
			name:            "link exists",
			id:              "MQ",
			wantOriginalURL: "https://example.com",
			wantFound:       true,
			wantStatusCode:  http.StatusTemporaryRedirect,
			wantLocation:    "https://example.com",
		},
		{
			name:            "link does not exist",
			id:              "unknown",
			wantOriginalURL: "",
			wantFound:       false,
			wantStatusCode:  http.StatusBadRequest,
			wantLocation:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewMockShortLinkService(t)

			service.EXPECT().
				GetSourceLink(
					mock.Anything,
					test.id,
				).
				Return(
					test.wantOriginalURL,
					test.wantFound,
					nil,
				).
				Once()

			router := NewRouter(service, "http://localhost:8080")

			request := httptest.NewRequest(
				http.MethodGet,
				"/"+test.id,
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

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
