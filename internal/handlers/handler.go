package handlers

import (
	"net/http"

	"github.com/alexia-23/shortener/internal/repository"
)

type Handler struct {
	repository *repository.ShortLinkRepository
}

func NewHandler(repo *repository.ShortLinkRepository) *Handler {
	return &Handler{
		repository: repo,
	}
}

func (handler *Handler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method == http.MethodGet {
		handler.handleRedirect(w, r)
		return
	}

	if r.Method == http.MethodPost {
		handler.handleCreateShortLink(w, r)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}
