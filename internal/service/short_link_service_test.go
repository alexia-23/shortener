package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepository struct {
	savedID          string
	savedOriginalURL string
	saveResults      []bool
	saveCalls        int
	links            map[string]string
}

func (mock *mockRepository) Save(
	id string,
	originalURL string,
) bool {
	mock.savedID = id
	mock.savedOriginalURL = originalURL
	mock.saveCalls++

	if len(mock.saveResults) == 0 {
		return false
	}

	result := mock.saveResults[0]
	mock.saveResults = mock.saveResults[1:]

	return result
}

func (mock *mockRepository) Get(
	id string,
) (string, bool) {
	originalURL, found := mock.links[id]

	return originalURL, found
}

func TestShortLinkService_CreateShortLink(t *testing.T) {
	repository := &mockRepository{
		saveResults: []bool{true},
		links:       make(map[string]string),
	}

	service := NewShortLinkService(repository)

	originalURL := "https://example.com"

	id, err := service.CreateShortLink(originalURL)

	require.NoError(t, err)
	require.NotEmpty(t, id)

	assert.Len(t, id, 8)
	assert.Equal(t, id, repository.savedID)
	assert.Equal(t, originalURL, repository.savedOriginalURL)
}

func TestShortLinkService_GetSourceLink(t *testing.T) {
	repository := &mockRepository{
		links: map[string]string{
			"test-id": "https://example.com",
		},
	}

	service := NewShortLinkService(repository)

	originalURL, found := service.GetSourceLink("test-id")

	require.True(t, found)
	assert.Equal(t, "https://example.com", originalURL)
}

func TestShortLinkService_CreateShortLink_RetriesOnCollision(
	t *testing.T,
) {
	repository := &mockRepository{
		saveResults: []bool{false, true},
		links:       make(map[string]string),
	}

	service := NewShortLinkService(repository)

	id, err := service.CreateShortLink(
		"https://example.com",
	)

	require.NoError(t, err)
	require.NotEmpty(t, id)

	assert.Equal(t, 2, repository.saveCalls)
}

func TestShortLinkService_CreateShortLink_ReturnsErrorAfterMaxAttempts(
	t *testing.T,
) {
	repository := &mockRepository{
		saveResults: make([]bool, maxGenerateAttempts),
		links:       make(map[string]string),
	}

	service := NewShortLinkService(repository)

	id, err := service.CreateShortLink(
		"https://example.com",
	)

	require.ErrorIs(t, err, ErrGenerateUniqueID)
	assert.Empty(t, id)
	assert.Equal(t, maxGenerateAttempts, repository.saveCalls)
}
