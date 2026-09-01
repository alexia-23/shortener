package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type createShortLinksBatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type createShortLinksBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (handler *Handler) handleCreateShortLinksBatch(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxRequestBodySize,
	)

	var request []createShortLinksBatchRequestItem

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeBadRequest(w)
		return
	}

	if len(request) == 0 {
		writeBadRequest(w)
		return
	}

	originalURLs := make([]string, 0, len(request))
	for index := range request {
		originalURL := strings.TrimSpace(request[index].OriginalURL)

		parsedURL, err := url.ParseRequestURI(originalURL)
		if originalURL == "" ||
			err != nil ||
			parsedURL.Scheme == "" ||
			parsedURL.Host == "" {
			writeBadRequest(w)
			return
		}

		request[index].OriginalURL = originalURL
		originalURLs = append(originalURLs, originalURL)
	}

	ids, err := handler.service.CreateShortLinksBatch(
		r.Context(),
		originalURLs,
	)
	if err != nil || len(ids) != len(request) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := make(
		[]createShortLinksBatchResponseItem,
		0,
		len(request),
	)

	for index, item := range request {
		shortURL, err := url.JoinPath(
			handler.baseURL,
			ids[index],
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response = append(
			response,
			createShortLinksBatchResponseItem{
				CorrelationID: item.CorrelationID,
				ShortURL:      shortURL,
			},
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(response)
}
