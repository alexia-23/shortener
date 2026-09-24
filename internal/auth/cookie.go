package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	userIDSize    = 16
	tokenLifetime = 30 * 24 * time.Hour
)

var ErrInvalidCookie = errors.New("invalid authentication cookie")

// Claims is the shared contract for issuing and verifying authentication tokens.
// An authenticated token without a UserID is handled as unauthorized by handlers.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type CookieSigner struct {
	secretKey []byte
}

func NewCookieSigner(secretKey string) *CookieSigner {
	if secretKey == "" {
		panic("auth: empty secret key")
	}

	return &CookieSigner{secretKey: []byte(secretKey)}
}

func GenerateUserID() (string, error) {
	randomBytes := make([]byte, userIDSize)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate user ID: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func (signer *CookieSigner) Sign(claims Claims) (string, error) {
	now := time.Now()

	if claims.IssuedAt == nil {
		claims.IssuedAt = jwt.NewNumericDate(now)
	}

	if claims.ExpiresAt == nil {
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(tokenLifetime))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	value, err := token.SignedString(signer.secretKey)
	if err != nil {
		return "", fmt.Errorf("sign authentication token: %w", err)
	}

	return value, nil
}

func (signer *CookieSigner) Verify(value string) (Claims, error) {
	var claims Claims

	token, err := jwt.ParseWithClaims(
		value,
		&claims,
		func(_ *jwt.Token) (any, error) {
			return signer.secretKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %v", ErrInvalidCookie, err)
	}

	if !token.Valid {
		return Claims{}, ErrInvalidCookie
	}

	return claims, nil
}
