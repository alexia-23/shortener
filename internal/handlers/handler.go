package handlers

import (
	"net/http"
	"sync"
)

var (
	links   = make(map[string]string)
	counter int
	mutex   sync.RWMutex
)

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handleRedirect(w, r)
		return
	}

	if r.Method == http.MethodPost {
		handleCreateShortLink(w, r)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}
