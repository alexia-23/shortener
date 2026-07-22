package handlers

type Handler struct {
	repository ShortLinkRepository
}

func NewHandler(repo ShortLinkRepository) *Handler {
	if repo == nil {
		panic("handlers: nil repository")
	}

	return &Handler{
		repository: repo,
	}
}
