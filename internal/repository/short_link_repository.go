package repository

import (
	"fmt"
	"sync"

	"github.com/alexia-23/shortener/internal/service"
)

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
) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if _, exists := repository.links[id]; exists {
		return fmt.Errorf(
			"short link with ID %q: %w",
			id,
			service.ErrShortLinkIDExists,
		)
	}

	repository.links[id] = originalURL

	return nil
}

func (repository *ShortLinkRepository) Get(
	id string,
) (string, bool) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	originalURL, found := repository.links[id]

	return originalURL, found
}
