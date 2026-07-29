package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShortLinkService_CreateShortLink(t *testing.T) {
	repository := NewMockShortLinkRepository(t)

	originalURL := "https://example.com"

	var savedID string

	repository.EXPECT().
		Save(mock.Anything, originalURL).
		Run(func(id string, originalURL string) {
			savedID = id
		}).
		Return(nil).
		Once()

	service := NewShortLinkService(repository)

	id, err := service.CreateShortLink(originalURL)

	require.NoError(t, err)
	require.NotEmpty(t, id)

	assert.Len(t, id, 8)
	assert.Equal(t, id, savedID)
}

func TestShortLinkService_GetSourceLink(t *testing.T) {
	repository := NewMockShortLinkRepository(t)

	repository.EXPECT().
		Get("test-id").
		Return("https://example.com", true).
		Once()

	service := NewShortLinkService(repository)

	originalURL, found := service.GetSourceLink("test-id")

	require.True(t, found)
	assert.Equal(t, "https://example.com", originalURL)
}

func TestShortLinkService_CreateShortLink_RetriesOnCollision(
	t *testing.T,
) {
	repository := NewMockShortLinkRepository(t)

	originalURL := "https://example.com"

	repository.EXPECT().
		Save(mock.Anything, originalURL).
		Return(ErrShortLinkIDExists).
		Once()

	repository.EXPECT().
		Save(mock.Anything, originalURL).
		Return(nil).
		Once()

	service := NewShortLinkService(repository)

	id, err := service.CreateShortLink(originalURL)

	require.NoError(t, err)
	require.NotEmpty(t, id)
	assert.Len(t, id, 8)
}

func TestShortLinkService_CreateShortLink_ReturnsErrorAfterMaxAttempts(
	t *testing.T,
) {
	repository := NewMockShortLinkRepository(t)

	originalURL := "https://example.com"

	repository.EXPECT().
		Save(mock.Anything, originalURL).
		Return(ErrShortLinkIDExists).
		Times(maxGenerateAttempts)

	service := NewShortLinkService(repository)

	id, err := service.CreateShortLink(originalURL)

	require.ErrorIs(t, err, ErrGenerateUniqueID)
	assert.Empty(t, id)
}
