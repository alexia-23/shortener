package handlers

import (
	"errors"
	"net/http"

	"github.com/alexia-23/shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func (handler *Handler) handleGetSourceLink(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")

	originalURL, found, err := handler.service.GetSourceLink(
		r.Context(),
		id,
	)

	if errors.Is(err, service.ErrShortLinkDeleted) {
		w.WriteHeader(http.StatusGone)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !found {
		writeBadRequest(w)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
