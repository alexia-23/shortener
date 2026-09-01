package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
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
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
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
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
