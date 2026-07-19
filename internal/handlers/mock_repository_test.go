package handlers

type mockRepository struct {
	saveID   string
	savedURL string
	links    map[string]string
}

func (repository *mockRepository) Save(originalURL string) string {
	repository.savedURL = originalURL

	return repository.saveID
}

func (repository *mockRepository) Get(id string) (string, bool) {
	originalURL, ok := repository.links[id]

	return originalURL, ok
}
