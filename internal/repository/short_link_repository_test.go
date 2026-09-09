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

func TestMemoryRepository_SaveBatch(t *testing.T) {
	repository := NewMemoryRepository()
	ctx := context.Background()

	links := []service.ShortLink{
		{
			ID:          "first-id",
			OriginalURL: "https://example.com/first",
		},
		{
			ID:          "second-id",
			OriginalURL: "https://example.com/second",
		},
	}

	require.NoError(t, repository.SaveBatch(ctx, links))

	for _, link := range links {
		originalURL, found, err := repository.Get(ctx, link.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, link.OriginalURL, originalURL)
	}
}

func TestMemoryRepository_SaveBatch_IsAtomicOnCollision(t *testing.T) {
	repository := NewMemoryRepository()
	ctx := context.Background()

	require.NoError(
		t,
		repository.Save(
			ctx,
			"existing-id",
			"https://example.com/existing",
		),
	)

	err := repository.SaveBatch(
		ctx,
		[]service.ShortLink{
			{
				ID:          "new-id",
				OriginalURL: "https://example.com/new",
			},
			{
				ID:          "existing-id",
				OriginalURL: "https://example.com/replacement",
			},
		},
	)

	require.ErrorIs(t, err, service.ErrShortLinkIDExists)

	_, found, getErr := repository.Get(ctx, "new-id")
	require.NoError(t, getErr)
	assert.False(t, found)

	originalURL, found, getErr := repository.Get(ctx, "existing-id")
	require.NoError(t, getErr)
	require.True(t, found)
	assert.Equal(t, "https://example.com/existing", originalURL)
}

func TestFileRepository_SaveBatchAndReload(t *testing.T) {
	fileStoragePath := filepath.Join(
		t.TempDir(),
		"storage.json",
	)
	ctx := context.Background()

	firstRepository, err := NewFileRepository(fileStoragePath)
	require.NoError(t, err)

	links := []service.ShortLink{
		{
			ID:          "first-id",
			OriginalURL: "https://example.com/first",
		},
		{
			ID:          "second-id",
			OriginalURL: "https://example.com/second",
		},
	}

	require.NoError(t, firstRepository.SaveBatch(ctx, links))

	secondRepository, err := NewFileRepository(fileStoragePath)
	require.NoError(t, err)

	for _, link := range links {
		originalURL, found, getErr := secondRepository.Get(
			ctx,
			link.ID,
		)
		require.NoError(t, getErr)
		require.True(t, found)
		assert.Equal(t, link.OriginalURL, originalURL)
	}
}

func TestFileRepository_SaveBatch_DuplicateDoesNotWrite(t *testing.T) {
	fileStoragePath := filepath.Join(
		t.TempDir(),
		"storage.json",
	)
	ctx := context.Background()

	repository, err := NewFileRepository(fileStoragePath)
	require.NoError(t, err)

	require.NoError(
		t,
		repository.Save(
			ctx,
			"existing-id",
			"https://example.com/existing",
		),
	)

	fileInfoBefore, err := os.Stat(fileStoragePath)
	require.NoError(t, err)

	err = repository.SaveBatch(
		ctx,
		[]service.ShortLink{
			{
				ID:          "new-id",
				OriginalURL: "https://example.com/new",
			},
			{
				ID:          "existing-id",
				OriginalURL: "https://example.com/replacement",
			},
		},
	)
	require.ErrorIs(t, err, service.ErrShortLinkIDExists)

	fileInfoAfter, err := os.Stat(fileStoragePath)
	require.NoError(t, err)
	assert.Equal(t, fileInfoBefore.Size(), fileInfoAfter.Size())

	_, found, getErr := repository.Get(ctx, "new-id")
	require.NoError(t, getErr)
	assert.False(t, found)
}
