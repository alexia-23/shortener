package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(repo ShortLinkRepository, baseURL string) *chi.Mux {
	handler := NewHandler(repo, baseURL)

	router := chi.NewRouter()

	router.Post("/", handler.handleCreateShortLink)
	router.Get("/{id}", handler.handleGetSourceLink)

	router.NotFound(badRequestHandler)
	router.MethodNotAllowed(badRequestHandler)

	return router
}

func badRequestHandler(w http.ResponseWriter, r *http.Request) {
	writeBadRequest(w)
}

func writeBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}
