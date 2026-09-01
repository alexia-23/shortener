package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/alexia-23/shortener/internal/service"
)

type createShortLinkRequest struct {
	URL string `json:"url"`
}

type createShortLinkResponse struct {
	Result string `json:"result"`
}

func (handler *Handler) handleCreateShortLinkJSON(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxRequestBodySize,
	)

	var request createShortLinkRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeBadRequest(w)
		return
	}

	originalURL, valid := validateOriginalURL(request.URL)
	if !valid {
		writeBadRequest(w)
		return
	}

	id, err := handler.service.CreateShortLink(
		r.Context(),
		originalURL,
	)
	statusCode := http.StatusCreated

	if err != nil {
		var originalURLExistsError *service.OriginalURLExistsError

		if !errors.As(err, &originalURLExistsError) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		id = originalURLExistsError.ID
		statusCode = http.StatusConflict
	}

	shortURL, err := url.JoinPath(handler.baseURL, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := createShortLinkResponse{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
