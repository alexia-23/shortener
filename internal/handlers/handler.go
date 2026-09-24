package handlers

type Handler struct {
	service URLService
	baseURL string
}

func NewHandler(service URLService, baseURL string) *Handler {
	if service == nil {
		panic("handlers: nil service")
	}

	return &Handler{
		service: service,
		baseURL: baseURL,
	}
}
