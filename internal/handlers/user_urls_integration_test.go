package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexia-23/shortener/internal/auth"
	"github.com/alexia-23/shortener/internal/middleware"
	"github.com/alexia-23/shortener/internal/repository"
	"github.com/alexia-23/shortener/internal/service"
)

func TestUserURLsFlow(t *testing.T) {
	memoryRepository := repository.NewMemoryRepository()
	shortLinkService := service.NewShortLinkService(
		memoryRepository,
	)

	signer := auth.NewCookieSigner(
		"test-secret-key",
	)

	router := NewRouter(
		shortLinkService,
		"http://localhost:8080",
		middleware.Authentication(signer),
	)

	const originalURL = "https://example.com/test"

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(originalURL),
	)
	createRequest.Header.Set(
		"Content-Type",
		"text/plain",
	)

	createRecorder := httptest.NewRecorder()

	router.ServeHTTP(
		createRecorder,
		createRequest,
	)

	createResponse := createRecorder.Result()
	defer createResponse.Body.Close()

	if createResponse.StatusCode != http.StatusCreated {
		t.Fatalf(
			"create status code = %d, want %d",
			createResponse.StatusCode,
			http.StatusCreated,
		)
	}

	cookies := createResponse.Cookies()
	if len(cookies) != 1 {
		t.Fatalf(
			"cookies count = %d, want 1",
			len(cookies),
		)
	}

	shortURL := strings.TrimSpace(
		createRecorder.Body.String(),
	)

	if shortURL == "" {
		t.Fatal("created short URL is empty")
	}

	getRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/user/urls",
		nil,
	)
	getRequest.AddCookie(cookies[0])

	getRecorder := httptest.NewRecorder()

	router.ServeHTTP(
		getRecorder,
		getRequest,
	)

	getResponse := getRecorder.Result()
	defer getResponse.Body.Close()

	if getResponse.StatusCode != http.StatusOK {
		t.Fatalf(
			"get status code = %d, want %d",
			getResponse.StatusCode,
			http.StatusOK,
		)
	}

	var response []userURLResponse

	if err := json.NewDecoder(
		getResponse.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"decode response: %v",
			err,
		)
	}

	if len(response) != 1 {
		t.Fatalf(
			"response length = %d, want 1",
			len(response),
		)
	}

	if response[0].OriginalURL != originalURL {
		t.Errorf(
			"original URL = %q, want %q",
			response[0].OriginalURL,
			originalURL,
		)
	}

	if response[0].ShortURL != shortURL {
		t.Errorf(
			"short URL = %q, want %q",
			response[0].ShortURL,
			shortURL,
		)
	}

	secondUserRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/user/urls",
		nil,
	)

	secondUserRecorder := httptest.NewRecorder()

	router.ServeHTTP(
		secondUserRecorder,
		secondUserRequest,
	)

	if secondUserRecorder.Code != http.StatusNoContent {
		t.Errorf(
			"second user status code = %d, want %d",
			secondUserRecorder.Code,
			http.StatusNoContent,
		)
	}
}

func TestUserURLsInvalidCookieIsReplaced(t *testing.T) {
	memoryRepository := repository.NewMemoryRepository()
	shortLinkService := service.NewShortLinkService(
		memoryRepository,
	)

	signer := auth.NewCookieSigner(
		"test-secret-key",
	)

	router := NewRouter(
		shortLinkService,
		"http://localhost:8080",
		middleware.Authentication(signer),
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/user/urls",
		nil,
	)

	request.AddCookie(
		&http.Cookie{
			Name:  middleware.UserCookieName,
			Value: "invalid-cookie",
		},
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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

	cookies := response.Cookies()
	if len(cookies) != 1 {
		t.Fatalf(
			"cookies count = %d, want 1",
			len(cookies),
		)
	}

	userID, err := signer.Verify(
		cookies[0].Value,
	)
	if err != nil {
		t.Fatalf(
			"verify new cookie: %v",
			err,
		)
	}

	if userID == "" {
		t.Fatal(
			"replacement cookie contains empty user ID",
		)
	}
}

func TestUserURLsValidCookieWithoutUserIDUnauthorized(
	t *testing.T,
) {
	memoryRepository := repository.NewMemoryRepository()
	shortLinkService := service.NewShortLinkService(
		memoryRepository,
	)

	signer := auth.NewCookieSigner(
		"test-secret-key",
	)

	router := NewRouter(
		shortLinkService,
		"http://localhost:8080",
		middleware.Authentication(signer),
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/user/urls",
		nil,
	)

	request.AddCookie(
		&http.Cookie{
			Name:  middleware.UserCookieName,
			Value: signer.Sign(""),
		},
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"status code = %d, want %d",
			recorder.Code,
			http.StatusUnauthorized,
		)
	}
}
