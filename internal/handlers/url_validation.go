package handlers

import (
	"net/url"
	"strings"
)

func validateOriginalURL(
	rawURL string,
) (string, bool) {
	originalURL := strings.TrimSpace(rawURL)

	if originalURL == "" {
		return "", false
	}

	parsedURL, err := url.ParseRequestURI(originalURL)
	if err != nil ||
		parsedURL.Scheme == "" ||
		parsedURL.Host == "" {
		return "", false
	}

	return originalURL, true
}
