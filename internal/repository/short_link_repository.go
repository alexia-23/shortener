package repository

import "sync"

type ShortLinkRepository struct {
	links map[string]string
	mutex sync.Mutex
}

func NewShortLinkRepository() *ShortLinkRepository {
	return &ShortLinkRepository{
		links: make(map[string]string),
	}
}

func (repository *ShortLinkRepository) Save(
	id string,
	originalURL string,
) bool {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if _, exists := repository.links[id]; exists {
		return false
	}

	repository.links[id] = originalURL

	return true
}

func (repository *ShortLinkRepository) Get(
	id string,
) (string, bool) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	originalURL, found := repository.links[id]

	return originalURL, found
}
