package handlers

type ShortLinkService interface {
	CreateShortLink(originalURL string) (string, error)
	GetSourceLink(id string) (string, bool)
}
