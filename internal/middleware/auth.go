package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alexia-23/shortener/internal/auth"
)

const UserCookieName = "user_id"

func Authentication(
	signer *auth.CookieSigner,
) func(http.Handler) http.Handler {
	if signer == nil {
		panic("middleware: nil cookie signer")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			userID, err := getOrCreateUserID(
				w,
				r,
				signer,
			)
			if err != nil {
				http.Error(
					w,
					"failed to authenticate user",
					http.StatusInternalServerError,
				)
				return
			}

			ctx := auth.ContextWithUserID(
				r.Context(),
				userID,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}

func getOrCreateUserID(
	w http.ResponseWriter,
	r *http.Request,
	signer *auth.CookieSigner,
) (string, error) {
	cookie, err := r.Cookie(UserCookieName)

	if err == nil {
		claims, verifyErr := signer.Verify(cookie.Value)
		if verifyErr == nil {
			return claims.UserID, nil
		}
	}

	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return "", fmt.Errorf(
			"read authentication cookie: %w",
			err,
		)
	}

	userID, err := auth.GenerateUserID()
	if err != nil {
		return "", fmt.Errorf(
			"generate user ID: %w",
			err,
		)
	}

	cookieValue, err := signer.Sign(
		auth.Claims{
			UserID: userID,
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"sign authentication cookie: %w",
			err,
		)
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name:     UserCookieName,
			Value:    cookieValue,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	)

	return userID, nil
}
