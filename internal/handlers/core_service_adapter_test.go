package handlers

import (
	"context"
	"net/http"

	"github.com/alexia-23/shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

// Core route tests must fail if they unexpectedly call a user operation.
// The generated MockShortLinkService remains a mock of the narrow interface.
type coreServiceAdapter struct {
	ShortLinkService
}

func (coreServiceAdapter) GetUserLinks(
	context.Context,
	string,
) ([]service.ShortLink, error) {
	panic("unexpected GetUserLinks call in a core route test")
}

func (coreServiceAdapter) DeleteUserLinks(
	string,
	[]string,
) {
	panic("unexpected DeleteUserLinks call in a core route test")
}

func newCoreTestRouter(
	service ShortLinkService,
	baseURL string,
	middlewares ...func(http.Handler) http.Handler,
) *chi.Mux {
	return NewRouter(
		coreServiceAdapter{
			ShortLinkService: service,
		},
		baseURL,
		middlewares...,
	)
}

var _ URLService = coreServiceAdapter{}
