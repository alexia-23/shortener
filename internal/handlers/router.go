package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	service ShortLinkService,
	baseURL string,
	middlewares ...func(http.Handler) http.Handler,
) *chi.Mux {
	handler := NewHandler(service, baseURL)

	router := chi.NewRouter()
	router.Use(middlewares...)

	router.Post("/", handler.handleCreateShortLink)
	router.Post("/api/shorten", handler.handleCreateShortLinkJSON)
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
