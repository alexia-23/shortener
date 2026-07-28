package handlers

type mockService struct {
	createID    string
	createErr   error
	createdURL  string
	sourceLinks map[string]string
}

func (mock *mockService) CreateShortLink(
	originalURL string,
) (string, error) {
	mock.createdURL = originalURL

	return mock.createID, mock.createErr
}

func (mock *mockService) GetSourceLink(
	id string,
) (string, bool) {
	originalURL, found := mock.sourceLinks[id]

	return originalURL, found
}
