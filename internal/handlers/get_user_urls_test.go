package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexia-23/shortener/internal/auth"
	"github.com/alexia-23/shortener/internal/service"
)

type userLinksServiceStub struct {
	links []service.ShortLink
	err   error
}

func (stub *userLinksServiceStub) CreateShortLink(
	_ context.Context,
	_ string,
) (string, error) {
	return "", nil
}

func (stub *userLinksServiceStub) CreateShortLinksBatch(
	_ context.Context,
	_ []string,
) ([]string, error) {
	return nil, nil
}

func (stub *userLinksServiceStub) GetSourceLink(
	_ context.Context,
	_ string,
) (string, bool, error) {
	return "", false, nil
}

func (stub *userLinksServiceStub) GetUserLinks(
	_ context.Context,
	_ string,
) ([]service.ShortLink, error) {
	return stub.links, stub.err
}

func TestHandleGetUserURLsUnauthorized(t *testing.T) {
	serviceStub := &userLinksServiceStub{}

	handler := NewHandler(
		serviceStub,
		"http://localhost:8080",
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/user/urls",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.handleGetUserURLs(
		recorder,
		request,
	)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"status code = %d, want %d",
			response.StatusCode,
			http.StatusUnauthorized,
		)
	}
}

func TestHandleGetUserURLsNoContent(t *testing.T) {
	serviceStub := &userLinksServiceStub{
		links: []service.ShortLink{},
	}

	handler := NewHandler(
		serviceStub,
		"http://localhost:8080",
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/user/urls",
		nil,
	)

	ctx := auth.ContextWithUserID(
		request.Context(),
		"user-1",
	)

	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()

	handler.handleGetUserURLs(
		recorder,
		request,
	)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf(
			"status code = %d, want %d",
			response.StatusCode,
			http.StatusNoContent,
		)
	}
}

func TestHandleGetUserURLsSuccess(t *testing.T) {
	serviceStub := &userLinksServiceStub{
		links: []service.ShortLink{
			{
				ID:          "abc123",
				OriginalURL: "https://example.com/one",
				UserID:      "user-1",
			},
			{
				ID:          "xyz789",
				OriginalURL: "https://example.com/two",
				UserID:      "user-1",
			},
		},
	}

	handler := NewHandler(
		serviceStub,
		"http://localhost:8080",
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/user/urls",
		nil,
	)

	ctx := auth.ContextWithUserID(
		request.Context(),
		"user-1",
	)

	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()

	handler.handleGetUserURLs(
		recorder,
		request,
	)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			response.StatusCode,
			http.StatusOK,
		)
	}

	if response.Header.Get("Content-Type") != "application/json" {
		t.Errorf(
			"Content-Type = %q, want %q",
			response.Header.Get("Content-Type"),
			"application/json",
		)
	}

	expectedBody := "" +
		"[{\"short_url\":\"http://localhost:8080/abc123\"," +
		"\"original_url\":\"https://example.com/one\"}," +
		"{\"short_url\":\"http://localhost:8080/xyz789\"," +
		"\"original_url\":\"https://example.com/two\"}]\n"

	if recorder.Body.String() != expectedBody {
		t.Errorf(
			"body = %q, want %q",
			recorder.Body.String(),
			expectedBody,
		)
	}
}

func TestHandleGetUserURLsServiceError(t *testing.T) {
	serviceStub := &userLinksServiceStub{
		err: errors.New("repository error"),
	}

	handler := NewHandler(
		serviceStub,
		"http://localhost:8080",
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/user/urls",
		nil,
	)

	ctx := auth.ContextWithUserID(
		request.Context(),
		"user-1",
	)

	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()

	handler.handleGetUserURLs(
		recorder,
		request,
	)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf(
			"status code = %d, want %d",
			response.StatusCode,
			http.StatusInternalServerError,
		)
	}
}
