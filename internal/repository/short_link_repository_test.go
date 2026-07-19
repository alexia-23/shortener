package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortLinkRepository_SaveAndGet(t *testing.T) {
	repository := NewShortLinkRepository()

	originalURL := "https://example.com"

	id := repository.Save(originalURL)

	require.NotEmpty(t, id)

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

func TestShortLinkRepository_SaveCreatesDifferentIDs(t *testing.T) {
	repository := NewShortLinkRepository()

	firstID := repository.Save("https://example.com/first")
	secondID := repository.Save("https://example.com/second")

	assert.NotEqual(t, firstID, secondID)
}

func TestCreateID(t *testing.T) {
	tests := []struct {
		name    string
		counter int
		want    string
	}{
		{
			name:    "counter one",
			counter: 1,
			want:    "MQ",
		},
		{
			name:    "counter two",
			counter: 2,
			want:    "Mg",
		},
		{
			name:    "counter ten",
			counter: 10,
			want:    "MTA",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := createID(test.counter)

			assert.Equal(t, test.want, result)
		})
	}
}
