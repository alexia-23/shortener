package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexia-23/shortener/internal/auth"
)

func TestAuthenticationCreatesCookie(t *testing.T) {
	signer := auth.NewCookieSigner("test-secret-key")

	var gotUserID string

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("user ID is missing from context")
		}

		gotUserID = userID
		w.WriteHeader(http.StatusOK)
	})

	handler := Authentication(signer)(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			response.StatusCode,
			http.StatusOK,
		)
	}

	if gotUserID == "" {
		t.Fatal("middleware passed empty user ID")
	}

	cookies := response.Cookies()
	if len(cookies) != 1 {
		t.Fatalf(
			"cookies count = %d, want 1",
			len(cookies),
		)
	}

	cookie := cookies[0]

	if cookie.Name != UserCookieName {
		t.Errorf(
			"cookie name = %q, want %q",
			cookie.Name,
			UserCookieName,
		)
	}

	claims, err := signer.Verify(cookie.Value)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if claims.UserID != gotUserID {
		t.Errorf(
			"cookie user ID = %q, context user ID = %q",
			claims.UserID,
			gotUserID,
		)
	}

	if !cookie.HttpOnly {
		t.Error("cookie HttpOnly = false, want true")
	}

	if cookie.Path != "/" {
		t.Errorf(
			"cookie Path = %q, want %q",
			cookie.Path,
			"/",
		)
	}
}

func TestAuthenticationUsesExistingValidCookie(t *testing.T) {
	signer := auth.NewCookieSigner("test-secret-key")

	const expectedUserID = "existing-user"

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("user ID is missing from context")
		}

		if userID != expectedUserID {
			t.Errorf(
				"user ID = %q, want %q",
				userID,
				expectedUserID,
			)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := Authentication(signer)(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	request.AddCookie(&http.Cookie{
		Name:  UserCookieName,
		Value: signedCookie(t, signer, expectedUserID),
	})

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			response.StatusCode,
			http.StatusOK,
		)
	}

	if len(response.Cookies()) != 0 {
		t.Errorf(
			"middleware unexpectedly replaced valid cookie",
		)
	}
}

func TestAuthenticationReplacesInvalidCookie(t *testing.T) {
	signer := auth.NewCookieSigner("test-secret-key")

	var gotUserID string

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("user ID is missing from context")
		}

		gotUserID = userID
		w.WriteHeader(http.StatusOK)
	})

	handler := Authentication(signer)(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	request.AddCookie(&http.Cookie{
		Name:  UserCookieName,
		Value: "invalid-cookie",
	})

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			response.StatusCode,
			http.StatusOK,
		)
	}

	cookies := response.Cookies()
	if len(cookies) != 1 {
		t.Fatalf(
			"cookies count = %d, want 1",
			len(cookies),
		)
	}

	newClaims, err := signer.Verify(cookies[0].Value)
	if err != nil {
		t.Fatalf(
			"Verify() new cookie error = %v",
			err,
		)
	}

	if newClaims.UserID == "" {
		t.Fatal("new cookie contains empty user ID")
	}

	if newClaims.UserID != gotUserID {
		t.Errorf(
			"cookie user ID = %q, context user ID = %q",
			newClaims.UserID,
			gotUserID,
		)
	}
}

func signedCookie(
	t *testing.T,
	signer *auth.CookieSigner,
	userID string,
) string {
	t.Helper()

	value, err := signer.Sign(
		auth.Claims{
			UserID: userID,
		},
	)
	if err != nil {
		t.Fatalf(
			"Sign() error = %v",
			err,
		)
	}

	return value
}
