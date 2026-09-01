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

func (repository *MemoryRepository) SaveBatch(
	_ context.Context,
	links []service.ShortLink,
) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	batchIDs := make(map[string]struct{}, len(links))

	for _, link := range links {
		if _, exists := repository.links[link.ID]; exists {
			return fmt.Errorf(
				"short link with ID %q: %w",
				link.ID,
				service.ErrShortLinkIDExists,
			)
		}

		if _, exists := batchIDs[link.ID]; exists {
			return fmt.Errorf(
				"duplicate short link ID %q in batch: %w",
				link.ID,
				service.ErrShortLinkIDExists,
			)
		}

		batchIDs[link.ID] = struct{}{}
	}

	for _, link := range links {
		repository.links[link.ID] = link.OriginalURL
	}

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

func (repository *MemoryRepository) deleteBatch(
	links []service.ShortLink,
) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	for _, link := range links {
		delete(repository.links, link.ID)
	}
}
