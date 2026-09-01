package handlers

import "context"

type ShortLinkService interface {
	CreateShortLink(
		ctx context.Context,
		originalURL string,
	) (string, error)

	GetSourceLink(
		ctx context.Context,
		id string,
	) (string, bool, error)
}
