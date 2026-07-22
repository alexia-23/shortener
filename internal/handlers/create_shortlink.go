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

	body, err := io.ReadAll(
		io.LimitReader(r.Body, maxRequestBodySize+1),
	)
	if err != nil || int64(len(body)) > maxRequestBodySize {
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

	id := handler.repository.Save(originalURL)
	shortURL := "http://localhost:8080/" + id

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, _ = w.Write([]byte(shortURL))
}
