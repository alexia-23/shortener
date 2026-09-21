package auth

import (
	"encoding/base64"
	"errors"
	"testing"
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

func TestCookieSignerSignAndVerify(t *testing.T) {
	signer := NewCookieSigner("test-secret-key")

	const userID = "test-user-id"

	cookieValue := signer.Sign(userID)

	gotUserID, err := signer.Verify(cookieValue)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if gotUserID != userID {
		t.Errorf(
			"Verify() user ID = %q, want %q",
			gotUserID,
			userID,
		)
	}
}

func TestCookieSignerVerifyInvalidCookie(t *testing.T) {
	signer := NewCookieSigner("test-secret-key")

	validCookie := signer.Sign("test-user-id")

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
			name:        "empty user ID with invalid signature",
			cookieValue: ".signature",
		},
		{
			name:        "invalid base64 signature",
			cookieValue: "test-user-id.%%%",
		},
		{
			name:        "changed user ID",
			cookieValue: "another-user" + validCookie[len("test-user-id"):],
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
				t.Errorf(
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

	cookieValue := firstSigner.Sign("test-user-id")

	_, err := secondSigner.Verify(cookieValue)
	if !errors.Is(err, ErrInvalidCookie) {
		t.Errorf(
			"Verify() error = %v, want %v",
			err,
			ErrInvalidCookie,
		)
	}
}

func TestCookieSignerVerifyValidCookieWithoutUserID(t *testing.T) {
	signer := NewCookieSigner("test-secret-key")

	cookieValue := signer.Sign("")

	userID, err := signer.Verify(cookieValue)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if userID != "" {
		t.Errorf(
			"Verify() user ID = %q, want empty string",
			userID,
		)
	}
}
