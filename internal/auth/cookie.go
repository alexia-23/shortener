package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

const userIDSize = 16

var ErrInvalidCookie = errors.New("invalid authentication cookie")

type CookieSigner struct {
	secretKey []byte
}

func NewCookieSigner(secretKey string) *CookieSigner {
	if secretKey == "" {
		panic("auth: empty secret key")
	}

	return &CookieSigner{
		secretKey: []byte(secretKey),
	}
}

func GenerateUserID() (string, error) {
	randomBytes := make([]byte, userIDSize)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate user ID: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func (signer *CookieSigner) Sign(userID string) string {
	signature := signer.makeSignature(userID)

	return userID + "." +
		base64.RawURLEncoding.EncodeToString(signature)
}

func (signer *CookieSigner) Verify(value string) (string, error) {
	userID, encodedSignature, found := strings.Cut(value, ".")
	if !found || encodedSignature == "" {
		return "", ErrInvalidCookie
	}

	signature, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return "", ErrInvalidCookie
	}

	expectedSignature := signer.makeSignature(userID)

	if !hmac.Equal(signature, expectedSignature) {
		return "", ErrInvalidCookie
	}

	return userID, nil
}

func (signer *CookieSigner) makeSignature(userID string) []byte {
	hash := hmac.New(sha256.New, signer.secretKey)

	_, _ = hash.Write([]byte(userID))

	return hash.Sum(nil)
}
