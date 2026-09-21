package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexia-23/shortener/internal/auth"
	"github.com/alexia-23/shortener/internal/service"
)

type MemoryRepository struct {
	links  map[string]string
	owners map[string]string
	mutex  sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		links:  make(map[string]string),
		owners: make(map[string]string),
	}
}

func (repository *MemoryRepository) Save(
	ctx context.Context,
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

	userID, _ := auth.UserIDFromContext(ctx)

	repository.links[id] = originalURL
	repository.owners[id] = userID

	return nil
}

func (repository *MemoryRepository) SaveBatch(
	ctx context.Context,
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

	userID, _ := auth.UserIDFromContext(ctx)

	for _, link := range links {
		repository.links[link.ID] = link.OriginalURL
		repository.owners[link.ID] = userID
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

func (repository *MemoryRepository) GetByUserID(
	_ context.Context,
	userID string,
) ([]service.ShortLink, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	links := make([]service.ShortLink, 0)

	for id, originalURL := range repository.links {
		if repository.owners[id] != userID {
			continue
		}

		links = append(
			links,
			service.ShortLink{
				ID:          id,
				OriginalURL: originalURL,
				UserID:      userID,
			},
		)
	}

	return links, nil
}

func (repository *MemoryRepository) deleteBatch(
	links []service.ShortLink,
) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	for _, link := range links {
		delete(repository.links, link.ID)
		delete(repository.owners, link.ID)
	}
}
