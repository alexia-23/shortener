package handlers

import (
	"net/http"
	"strings"
)

func (handler *Handler) handleRedirect(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := strings.TrimPrefix(r.URL.Path, "/")

	if id == "" || strings.Contains(id, "/") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, ok := handler.repository.Get(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
