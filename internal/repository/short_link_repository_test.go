package repository

import (
	"bufio"
	"context"
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

	ctx := context.Background()

	id := "test-id"
	originalURL := "https://example.com"

	err = repository.Save(
		ctx,
		id,
		originalURL,
	)
	require.NoError(t, err)

	savedURL, found, err := repository.Get(
		ctx,
		id,
	)
	require.NoError(t, err)

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

	ctx := context.Background()

	originalURL, found, err := repository.Get(
		ctx,
		"unknown",
	)
	require.NoError(t, err)

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

	ctx := context.Background()

	firstErr := repository.Save(
		ctx,
		"first-id",
		"https://example.com/first",
	)

	secondErr := repository.Save(
		ctx,
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

	ctx := context.Background()

	id := "same-id"

	firstErr := repository.Save(
		ctx,
		id,
		"https://example.com/first",
	)

	secondErr := repository.Save(
		ctx,
		id,
		"https://example.com/second",
	)

	require.NoError(t, firstErr)
	require.ErrorIs(
		t,
		secondErr,
		service.ErrShortLinkIDExists,
	)

	savedURL, found, err := repository.Get(
		ctx,
		id,
	)
	require.NoError(t, err)

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

	ctx := context.Background()

	id := "test-id"
	originalURL := "https://example.com"

	err = repository.Save(
		ctx,
		id,
		originalURL,
	)
	require.NoError(t, err)

	file, err := os.Open(fileStoragePath)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	require.True(t, scanner.Scan())
	require.NoError(t, scanner.Err())

	var savedURL storedURL

	err = json.Unmarshal(scanner.Bytes(), &savedURL)
	require.NoError(t, err)

	assert.Equal(t, "1", savedURL.UUID)
	assert.Equal(t, id, savedURL.ShortURL)
	assert.Equal(t, originalURL, savedURL.OriginalURL)

	assert.False(t, scanner.Scan())
	require.NoError(t, scanner.Err())
}

func TestShortLinkRepository_AppendToFile(t *testing.T) {
	fileStoragePath := filepath.Join(
		t.TempDir(),
		"storage.json",
	)

	repository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	ctx := context.Background()

	err = repository.Save(
		ctx,
		"first-id",
		"https://example.com/first",
	)
	require.NoError(t, err)

	err = repository.Save(
		ctx,
		"second-id",
		"https://example.com/second",
	)
	require.NoError(t, err)

	file, err := os.Open(fileStoragePath)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var records []storedURL

	for scanner.Scan() {
		var record storedURL

		err = json.Unmarshal(scanner.Bytes(), &record)
		require.NoError(t, err)

		records = append(records, record)
	}

	require.NoError(t, scanner.Err())
	require.Len(t, records, 2)

	assert.Equal(t, "1", records[0].UUID)
	assert.Equal(t, "first-id", records[0].ShortURL)
	assert.Equal(
		t,
		"https://example.com/first",
		records[0].OriginalURL,
	)

	assert.Equal(t, "2", records[1].UUID)
	assert.Equal(t, "second-id", records[1].ShortURL)
	assert.Equal(
		t,
		"https://example.com/second",
		records[1].OriginalURL,
	)
}

func TestShortLinkRepository_LoadFromFile(t *testing.T) {
	fileStoragePath := filepath.Join(
		t.TempDir(),
		"storage.json",
	)

	firstRepository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	ctx := context.Background()

	id := "test-id"
	originalURL := "https://example.com"

	err = firstRepository.Save(
		ctx,
		id,
		originalURL,
	)
	require.NoError(t, err)

	secondRepository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	savedURL, found, err := secondRepository.Get(
		ctx,
		id,
	)
	require.NoError(t, err)

	require.True(t, found)
	assert.Equal(t, originalURL, savedURL)
}

func TestShortLinkRepository_ContinuesUUIDAfterReload(t *testing.T) {
	fileStoragePath := filepath.Join(
		t.TempDir(),
		"storage.json",
	)

	firstRepository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	ctx := context.Background()

	err = firstRepository.Save(
		ctx,
		"first-id",
		"https://example.com/first",
	)
	require.NoError(t, err)

	secondRepository, err := NewShortLinkRepository(fileStoragePath)
	require.NoError(t, err)

	err = secondRepository.Save(
		ctx,
		"second-id",
		"https://example.com/second",
	)
	require.NoError(t, err)

	file, err := os.Open(fileStoragePath)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var records []storedURL

	for scanner.Scan() {
		var record storedURL

		err = json.Unmarshal(scanner.Bytes(), &record)
		require.NoError(t, err)

		records = append(records, record)
	}

	require.NoError(t, scanner.Err())
	require.Len(t, records, 2)

	assert.Equal(t, "1", records[0].UUID)
	assert.Equal(t, "2", records[1].UUID)
}
