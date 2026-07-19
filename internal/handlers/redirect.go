package handlers

import (
	"net/http"
	"strings"
)

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")

	if id == "" || strings.Contains(id, "/") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mutex.RLock()
	originalURL, ok := links[id]
	mutex.RUnlock()

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
