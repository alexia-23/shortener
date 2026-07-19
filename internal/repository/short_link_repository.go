package repository

import (
	"encoding/base64"
	"strconv"
	"sync"
)

type ShortLinkRepository struct {
	links   map[string]string
	counter int
	mutex   sync.RWMutex
}

func NewShortLinkRepository() *ShortLinkRepository {
	return &ShortLinkRepository{
		links: make(map[string]string),
	}
}

func (repository *ShortLinkRepository) Save(originalURL string) string {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	repository.counter++

	id := createID(repository.counter)
	repository.links[id] = originalURL

	return id
}

func (repository *ShortLinkRepository) Get(id string) (string, bool) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	originalURL, found := repository.links[id]

	return originalURL, found
}

func createID(counter int) string {
	number := strconv.Itoa(counter)

	return base64.RawURLEncoding.EncodeToString([]byte(number))
}
