package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexia-23/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortLinkRepository_SaveAndGet(t *testing.T) {
	fileStoragePath := filepath.Join(
		t.TempDir(),
		"storage.json",
	)

	repository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	id := "test-id"
	originalURL := "https://example.com"

	err = repository.Save(id, originalURL)
	require.NoError(t, err)

	savedURL, found := repository.Get(id)

	assert.True(t, found)
	assert.Equal(t, originalURL, savedURL)
}

func TestShortLinkRepository_GetUnknownID(t *testing.T) {
	fileStoragePath := filepath.Join(
		t.TempDir(),
		"storage.json",
	)

	repository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	originalURL, found := repository.Get("unknown")

	assert.False(t, found)
	assert.Empty(t, originalURL)
}

func TestShortLinkRepository_SaveDifferentIDs(t *testing.T) {
	fileStoragePath := filepath.Join(
		t.TempDir(),
		"storage.json",
	)

	repository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	firstErr := repository.Save(
		"first-id",
		"https://example.com/first",
	)

	secondErr := repository.Save(
		"second-id",
		"https://example.com/second",
	)

	require.NoError(t, firstErr)
	require.NoError(t, secondErr)
}

func TestShortLinkRepository_SaveCollision(t *testing.T) {
	fileStoragePath := filepath.Join(
		t.TempDir(),
		"storage.json",
	)

	repository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	id := "same-id"

	firstErr := repository.Save(
		id,
		"https://example.com/first",
	)

	secondErr := repository.Save(
		id,
		"https://example.com/second",
	)

	require.NoError(t, firstErr)
	require.ErrorIs(
		t,
		secondErr,
		service.ErrShortLinkIDExists,
	)

	savedURL, found := repository.Get(id)

	require.True(t, found)
	assert.Equal(
		t,
		"https://example.com/first",
		savedURL,
	)
}

func TestShortLinkRepository_SaveToFile(t *testing.T) {
	fileStoragePath := filepath.Join(
		t.TempDir(),
		"storage.json",
	)

	repository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	id := "test-id"
	originalURL := "https://example.com"

	err = repository.Save(id, originalURL)
	require.NoError(t, err)

	data, err := os.ReadFile(fileStoragePath)
	require.NoError(t, err)

	var urls []storedURL

	err = json.Unmarshal(data, &urls)
	require.NoError(t, err)

	require.Len(t, urls, 1)

	assert.Equal(t, "1", urls[0].UUID)
	assert.Equal(t, id, urls[0].ShortURL)
	assert.Equal(t, originalURL, urls[0].OriginalURL)
}

func TestShortLinkRepository_LoadFromFile(t *testing.T) {
	fileStoragePath := filepath.Join(
		t.TempDir(),
		"storage.json",
	)

	firstRepository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	id := "test-id"
	originalURL := "https://example.com"

	err = firstRepository.Save(id, originalURL)
	require.NoError(t, err)

	secondRepository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	savedURL, found := secondRepository.Get(id)

	require.True(t, found)
	assert.Equal(t, originalURL, savedURL)
}
