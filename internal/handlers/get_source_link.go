package handlers

import (
	"net/http"

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
