package handlers

import "net/http"

type Handler struct {
	repository ShortLinkRepository
}

func NewHandler(repo ShortLinkRepository) *Handler {
	return &Handler{
		repository: repo,
	}
}

func (handler *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /{$}", handler.handleCreateShortLink)
	mux.HandleFunc("GET /{id}", handler.handleGetSourceLink)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
}
