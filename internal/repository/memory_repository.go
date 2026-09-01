package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexia-23/shortener/internal/service"
)

type MemoryRepository struct {
	links map[string]string
	mutex sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		links: make(map[string]string),
	}
}

func (repository *MemoryRepository) Save(
	_ context.Context,
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

func (repository *MemoryRepository) Get(
	_ context.Context,
	id string,
) (string, bool, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	originalURL, found := repository.links[id]

	return originalURL, found, nil
}

func (repository *MemoryRepository) delete(id string) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	delete(repository.links, id)
}
