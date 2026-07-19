package handlers

type ShortLinkRepository interface {
	Save(originalURL string) string
	Get(id string) (string, bool)
}
