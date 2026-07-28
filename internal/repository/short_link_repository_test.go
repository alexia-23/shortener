package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortLinkRepository_SaveAndGet(t *testing.T) {
	repository := NewShortLinkRepository()

	id := "test-id"
	originalURL := "https://example.com"

	saved := repository.Save(id, originalURL)

	require.True(t, saved)

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

	firstSaved := repository.Save(
		"first-id",
		"https://example.com/first",
	)

	secondSaved := repository.Save(
		"second-id",
		"https://example.com/second",
	)

	assert.True(t, firstSaved)
	assert.True(t, secondSaved)
}

func TestShortLinkRepository_SaveCollision(t *testing.T) {
	repository := NewShortLinkRepository()

	id := "same-id"

	firstSaved := repository.Save(
		id,
		"https://example.com/first",
	)

	secondSaved := repository.Save(
		id,
		"https://example.com/second",
	)

	require.True(t, firstSaved)
	require.False(t, secondSaved)

	savedURL, found := repository.Get(id)

	require.True(t, found)
	assert.Equal(
		t,
		"https://example.com/first",
		savedURL,
	)
}
