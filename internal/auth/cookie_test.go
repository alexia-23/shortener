package auth

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateUserID(t *testing.T) {
	userID, err := GenerateUserID()
	if err != nil {
		t.Fatalf("GenerateUserID() error = %v", err)
	}

	if userID == "" {
		t.Fatal("GenerateUserID() returned empty user ID")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(userID)
	if err != nil {
		t.Fatalf("decode generated user ID: %v", err)
	}

	if len(decoded) != userIDSize {
		t.Fatalf(
			"generated user ID size = %d, want %d",
			len(decoded),
			userIDSize,
		)
	}
}

func mustSignClaims(t *testing.T, signer *CookieSigner, claims Claims) string {
	t.Helper()

	value, err := signer.Sign(claims)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	return value
}

func TestCookieSignerSignAndVerify(t *testing.T) {
	signer := NewCookieSigner("test-secret-key")

	want := Claims{
		UserID: "test-user-id",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "anonymous-user",
			Issuer:  "shortener",
		},
	}

	value := mustSignClaims(t, signer, want)

	if len(strings.Split(value, ".")) != 3 {
		t.Fatalf("expected a three-part JWT, got %q", value)
	}

	got, err := signer.Verify(value)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if got.UserID != want.UserID ||
		got.Subject != want.Subject ||
		got.Issuer != want.Issuer {
		t.Fatalf("Verify() claims = %+v, want fields from %+v", got, want)
	}

	if got.IssuedAt == nil ||
		got.ExpiresAt == nil ||
		!got.ExpiresAt.After(got.IssuedAt.Time) {
		t.Fatalf("JWT lifetime not set correctly: %+v", got.RegisteredClaims)
	}
}

func TestCookieSignerVerifyInvalidCookie(t *testing.T) {
	signer := NewCookieSigner("test-secret-key")

	validCookie := mustSignClaims(
		t,
		signer,
		Claims{UserID: "test-user-id"},
	)

	parts := strings.Split(validCookie, ".")

	changedPayload := parts[0] + "." +
		base64.RawURLEncoding.EncodeToString(
			[]byte(`{"user_id":"another-user","exp":4102444800}`),
		) +
		"." + parts[2]

	tests := []struct {
		name        string
		cookieValue string
	}{
		{
			name:        "empty cookie",
			cookieValue: "",
		},
		{
			name:        "missing signature",
			cookieValue: "test-user-id",
		},
		{
			name:        "legacy custom cookie",
			cookieValue: "test-user-id.signature",
		},
		{
			name: "invalid base64 signature",
			cookieValue: parts[0] +
				"." +
				parts[1] +
				".%%%",
		},
		{
			name:        "changed user ID",
			cookieValue: changedPayload,
		},
		{
			name:        "changed signature",
			cookieValue: validCookie + "x",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := signer.Verify(test.cookieValue)

			if !errors.Is(err, ErrInvalidCookie) {
				t.Fatalf(
					"Verify() error = %v, want %v",
					err,
					ErrInvalidCookie,
				)
			}
		})
	}
}

func TestCookieSignerDifferentSecret(t *testing.T) {
	firstSigner := NewCookieSigner("first-secret")
	secondSigner := NewCookieSigner("second-secret")

	value := mustSignClaims(
		t,
		firstSigner,
		Claims{UserID: "test-user-id"},
	)

	_, err := secondSigner.Verify(value)
	if !errors.Is(err, ErrInvalidCookie) {
		t.Fatalf(
			"Verify() error = %v, want %v",
			err,
			ErrInvalidCookie,
		)
	}
}

func TestCookieSignerVerifyValidCookieWithoutUserID(t *testing.T) {
	signer := NewCookieSigner("test-secret-key")

	value := mustSignClaims(t, signer, Claims{})

	claims, err := signer.Verify(value)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if claims.UserID != "" {
		t.Fatalf(
			"Verify() UserID = %q, want empty string",
			claims.UserID,
		)
	}
}

func TestCookieSignerRejectsInvalidRegisteredClaims(t *testing.T) {
	signer := NewCookieSigner("test-secret-key")
	now := time.Now()

	tests := []struct {
		name       string
		registered jwt.RegisteredClaims
	}{
		{
			name: "expired",
			registered: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(
					now.Add(-time.Minute),
				),
			},
		},
		{
			name: "not active yet",
			registered: jwt.RegisteredClaims{
				NotBefore: jwt.NewNumericDate(
					now.Add(time.Hour),
				),
			},
		},
		{
			name: "issued in future",
			registered: jwt.RegisteredClaims{
				IssuedAt: jwt.NewNumericDate(
					now.Add(time.Hour),
				),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := mustSignClaims(
				t,
				signer,
				Claims{
					UserID:           "user",
					RegisteredClaims: test.registered,
				},
			)

			if _, err := signer.Verify(value); !errors.Is(
				err,
				ErrInvalidCookie,
			) {
				t.Fatalf(
					"Verify() error = %v, want ErrInvalidCookie",
					err,
				)
			}
		})
	}
}

func TestCookieSignerRejectsMissingExpiration(t *testing.T) {
	const secret = "test-secret-key"

	signer := NewCookieSigner(secret)

	// Bypass Sign, which supplies an expiry for tokens issued by our app.
	value, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{UserID: "user"},
	).SignedString([]byte(secret))

	if err != nil {
		t.Fatal(err)
	}

	if _, err := signer.Verify(value); !errors.Is(
		err,
		ErrInvalidCookie,
	) {
		t.Fatalf(
			"Verify() error = %v, want ErrInvalidCookie",
			err,
		)
	}
}

func TestCookieSignerRejectsUnexpectedAlgorithm(t *testing.T) {
	const secret = "test-secret-key"

	signer := NewCookieSigner(secret)

	tests := []struct {
		name   string
		method jwt.SigningMethod
		key    any
	}{
		{
			name:   "HS384",
			method: jwt.SigningMethodHS384,
			key:    []byte(secret),
		},
		{
			name:   "none",
			method: jwt.SigningMethodNone,
			key:    jwt.UnsafeAllowNoneSignatureType,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := Claims{
				UserID: "user",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(
						time.Now().Add(time.Hour),
					),
				},
			}

			value, err := jwt.NewWithClaims(
				test.method,
				claims,
			).SignedString(test.key)

			if err != nil {
				t.Fatal(err)
			}

			if _, err := signer.Verify(value); !errors.Is(
				err,
				ErrInvalidCookie,
			) {
				t.Fatalf(
					"Verify() error = %v, want ErrInvalidCookie",
					err,
				)
			}
		})
	}
}

func TestCookieSignerPreservesExplicitExpiration(t *testing.T) {
	signer := NewCookieSigner("test-secret-key")

	expiration := jwt.NewNumericDate(
		time.Now().Add(time.Hour),
	)

	value := mustSignClaims(
		t,
		signer,
		Claims{
			UserID: "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: expiration,
			},
		},
	)

	got, err := signer.Verify(value)
	if err != nil {
		t.Fatal(err)
	}

	if got.ExpiresAt == nil ||
		!got.ExpiresAt.Equal(expiration.Time) {
		t.Fatalf(
			"expiration was changed: got %v, want %v",
			got.ExpiresAt,
			expiration,
		)
	}
}
