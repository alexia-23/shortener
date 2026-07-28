package handlers

type Handler struct {
	repository ShortLinkRepository
	baseURL    string
}

func NewHandler(repo ShortLinkRepository, baseURL string) *Handler {
	if repo == nil {
		panic("handlers: nil repository")
	}

	return &Handler{
		repository: repo,
		baseURL:    baseURL,
	}
}
