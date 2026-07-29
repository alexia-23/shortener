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

	originalURL, ok := handler.service.GetSourceLink(id)
	if !ok {
		writeBadRequest(w)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
