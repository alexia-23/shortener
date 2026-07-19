package handlers

import "net/http"

func (handler *Handler) handleGetSourceLink(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	originalURL, ok := handler.repository.Get(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
