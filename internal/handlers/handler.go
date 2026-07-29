package handlers

type Handler struct {
	service ShortLinkService
	baseURL string
}

func NewHandler(service ShortLinkService, baseURL string) *Handler {
	if service == nil {
		panic("handlers: nil service")
	}

	return &Handler{
		service: service,
		baseURL: baseURL,
	}
}
