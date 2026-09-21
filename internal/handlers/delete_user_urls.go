package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/alexia-23/shortener/internal/auth"
)

func (handler *Handler) handleDeleteUserURLs(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxRequestBodySize,
	)

	var ids []string

	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		writeBadRequest(w)
		return
	}

	deleteService, ok := handler.service.(DeleteUserLinksService)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	deleteService.DeleteUserLinks(
		userID,
		ids,
	)

	w.WriteHeader(http.StatusAccepted)
}
