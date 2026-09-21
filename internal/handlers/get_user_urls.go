package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/alexia-23/shortener/internal/auth"
)

type userURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (handler *Handler) handleGetUserURLs(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userLinksService, ok := handler.service.(UserLinksService)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	links, err := userLinksService.GetUserLinks(
		r.Context(),
		userID,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(links) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make(
		[]userURLResponse,
		0,
		len(links),
	)

	for _, link := range links {
		shortURL, err := url.JoinPath(
			handler.baseURL,
			link.ID,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response = append(
			response,
			userURLResponse{
				ShortURL:    shortURL,
				OriginalURL: link.OriginalURL,
			},
		)
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
