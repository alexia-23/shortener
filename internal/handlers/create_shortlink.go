package handlers

import (
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

const maxRequestBodySize int64 = 1 << 20

func (handler *Handler) handleCreateShortLink(
	w http.ResponseWriter,
	r *http.Request,
) {

	contentType := r.Header.Get("Content-Type")

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "text/plain" {
		writeBadRequest(w)
		return
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxRequestBodySize,
	)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeBadRequest(w)
		return
	}

	originalURL := strings.TrimSpace(string(body))

	if originalURL == "" {
		writeBadRequest(w)
		return
	}

	parsedURL, err := url.ParseRequestURI(originalURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		writeBadRequest(w)
		return
	}

	id, err := handler.service.CreateShortLink(originalURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortURL, err := url.JoinPath(handler.baseURL, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, _ = w.Write([]byte(shortURL))
}
