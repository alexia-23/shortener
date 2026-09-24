package handlers

import (
	"context"

	"github.com/alexia-23/shortener/internal/service"
)

type ShortLinkService interface {
	CreateShortLink(
		ctx context.Context,
		originalURL string,
	) (string, error)

	CreateShortLinksBatch(
		ctx context.Context,
		originalURLs []string,
	) ([]string, error)

	GetSourceLink(
		ctx context.Context,
		id string,
	) (string, bool, error)
}

type DeleteUserLinksService interface {
	DeleteUserLinks(
		userID string,
		ids []string,
	)
}

type UserLinksService interface {
	GetUserLinks(
		ctx context.Context,
		userID string,
	) ([]service.ShortLink, error)
}

// URLService contains every capability required by the HTTP router.
type URLService interface {
	ShortLinkService
	UserLinksService
	DeleteUserLinksService
}
