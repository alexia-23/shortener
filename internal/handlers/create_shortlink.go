package handlers

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"

	"github.com/alexia-23/shortener/internal/service"
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

	originalURL, valid := validateOriginalURL(string(body))
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

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)

	_, _ = w.Write([]byte(shortURL))
}
