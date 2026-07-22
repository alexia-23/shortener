package repository

import (
	"encoding/base64"
	"strconv"
	"sync"
)

type ShortLinkRepository struct {
	links   map[string]string
	counter uint64
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

func createID(counter uint64) string {
	number := strconv.FormatUint(counter, 10)

	return base64.RawURLEncoding.EncodeToString([]byte(number))
}
