package repository

import (
	"testing"

	"github.com/alexia-23/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortLinkRepository_SaveAndGet(t *testing.T) {
	repository := NewShortLinkRepository()

	id := "test-id"
	originalURL := "https://example.com"

	err := repository.Save(id, originalURL)

	require.NoError(t, err)

	savedURL, found := repository.Get(id)

	assert.True(t, found)
	assert.Equal(t, originalURL, savedURL)
}

func TestShortLinkRepository_GetUnknownID(t *testing.T) {
	repository := NewShortLinkRepository()

	originalURL, found := repository.Get("unknown")

	assert.False(t, found)
	assert.Empty(t, originalURL)
}

func TestShortLinkRepository_SaveDifferentIDs(t *testing.T) {
	repository := NewShortLinkRepository()

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
	repository := NewShortLinkRepository()

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
