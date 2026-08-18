package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/alexia-23/shortener/internal/service"
)

type ShortLinkRepository struct {
	links           map[string]string
	records         []storedURL
	mutex           sync.Mutex
	fileStoragePath string
}

type storedURL struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewShortLinkRepository(
	fileStoragePath string,
) (*ShortLinkRepository, error) {
	repository := &ShortLinkRepository{
		links:           make(map[string]string),
		records:         make([]storedURL, 0),
		fileStoragePath: fileStoragePath,
	}

	if err := repository.loadFromFile(); err != nil {
		return nil, err
	}

	return repository, nil
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

	record := storedURL{
		UUID:        strconv.Itoa(len(repository.records) + 1),
		ShortURL:    id,
		OriginalURL: originalURL,
	}

	repository.links[id] = originalURL
	repository.records = append(repository.records, record)

	if err := repository.saveToFile(); err != nil {
		delete(repository.links, id)
		repository.records = repository.records[:len(repository.records)-1]

		return err
	}

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

func (repository *ShortLinkRepository) saveToFile() error {
	data, err := json.MarshalIndent(
		repository.records,
		"",
		"  ",
	)
	if err != nil {
		return fmt.Errorf("marshal stored URLs: %w", err)
	}

	if err := os.WriteFile(
		repository.fileStoragePath,
		data,
		0666,
	); err != nil {
		return fmt.Errorf("write storage file: %w", err)
	}

	return nil
}

func (repository *ShortLinkRepository) loadFromFile() error {
	data, err := os.ReadFile(repository.fileStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("read storage file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var urls []storedURL

	if err := json.Unmarshal(data, &urls); err != nil {
		return fmt.Errorf("unmarshal stored URLs: %w", err)
	}

	repository.records = urls

	for _, storedURL := range urls {
		repository.links[storedURL.ShortURL] = storedURL.OriginalURL
	}

	return nil
}
