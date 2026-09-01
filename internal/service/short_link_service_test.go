package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShortLinkService_CreateShortLink(t *testing.T) {
	repository := NewMockShortLinkRepository(t)

	ctx := context.Background()
	originalURL := "https://example.com"

	var savedID string

	repository.EXPECT().
		Save(
			mock.Anything,
			mock.Anything,
			originalURL,
		).
		Run(func(
			ctx context.Context,
			id string,
			originalURL string,
		) {
			savedID = id
		}).
		Return(nil).
		Once()

	service := NewShortLinkService(repository)

	id, err := service.CreateShortLink(
		ctx,
		originalURL,
	)

	require.NoError(t, err)
	require.NotEmpty(t, id)

	assert.Len(t, id, 8)
	assert.Equal(t, id, savedID)
}

func TestShortLinkService_GetSourceLink(t *testing.T) {
	repository := NewMockShortLinkRepository(t)

	ctx := context.Background()

	repository.EXPECT().
		Get(
			mock.Anything,
			"test-id",
		).
		Return(
			"https://example.com",
			true,
			nil,
		).
		Once()

	service := NewShortLinkService(repository)

	originalURL, found, err := service.GetSourceLink(
		ctx,
		"test-id",
	)

	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "https://example.com", originalURL)
}

func TestShortLinkService_CreateShortLink_RetriesOnCollision(
	t *testing.T,
) {
	repository := NewMockShortLinkRepository(t)

	ctx := context.Background()
	originalURL := "https://example.com"

	repository.EXPECT().
		Save(
			mock.Anything,
			mock.Anything,
			originalURL,
		).
		Return(ErrShortLinkIDExists).
		Once()

	repository.EXPECT().
		Save(
			mock.Anything,
			mock.Anything,
			originalURL,
		).
		Return(nil).
		Once()

	service := NewShortLinkService(repository)

	id, err := service.CreateShortLink(
		ctx,
		originalURL,
	)

	require.NoError(t, err)
	require.NotEmpty(t, id)
	assert.Len(t, id, 8)
}

func TestShortLinkService_CreateShortLink_ReturnsErrorAfterMaxAttempts(
	t *testing.T,
) {
	repository := NewMockShortLinkRepository(t)

	ctx := context.Background()
	originalURL := "https://example.com"

	repository.EXPECT().
		Save(
			mock.Anything,
			mock.Anything,
			originalURL,
		).
		Return(ErrShortLinkIDExists).
		Times(maxGenerateAttempts)

	service := NewShortLinkService(repository)

	id, err := service.CreateShortLink(
		ctx,
		originalURL,
	)

	require.ErrorIs(t, err, ErrGenerateUniqueID)
	assert.Empty(t, id)
}

func TestShortLinkService_CreateShortLinksBatch(t *testing.T) {
	repository := NewMockShortLinkRepository(t)

	originalURLs := []string{
		"https://example.com/first",
		"https://example.com/second",
	}

	var savedLinks []ShortLink

	repository.EXPECT().
		SaveBatch(
			mock.Anything,
			mock.Anything,
		).
		Run(func(_ context.Context, links []ShortLink) {
			savedLinks = append(savedLinks, links...)
		}).
		Return(nil).
		Once()

	shortLinkService := NewShortLinkService(repository)

	ids, err := shortLinkService.CreateShortLinksBatch(
		context.Background(),
		originalURLs,
	)

	require.NoError(t, err)
	require.Len(t, ids, len(originalURLs))
	require.Len(t, savedLinks, len(originalURLs))

	for index := range originalURLs {
		assert.Len(t, ids[index], 8)
		assert.Equal(t, ids[index], savedLinks[index].ID)
		assert.Equal(
			t,
			originalURLs[index],
			savedLinks[index].OriginalURL,
		)
	}

	assert.NotEqual(t, ids[0], ids[1])
}

func TestShortLinkService_CreateShortLinksBatch_RetriesCollision(
	t *testing.T,
) {
	repository := NewMockShortLinkRepository(t)

	repository.EXPECT().
		SaveBatch(mock.Anything, mock.Anything).
		Return(ErrShortLinkIDExists).
		Once()

	repository.EXPECT().
		SaveBatch(mock.Anything, mock.Anything).
		Return(nil).
		Once()

	shortLinkService := NewShortLinkService(repository)

	ids, err := shortLinkService.CreateShortLinksBatch(
		context.Background(),
		[]string{"https://example.com"},
	)

	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Len(t, ids[0], 8)
}

func TestShortLinkService_CreateShortLinksBatch_Empty(t *testing.T) {
	repository := NewMockShortLinkRepository(t)
	shortLinkService := NewShortLinkService(repository)

	ids, err := shortLinkService.CreateShortLinksBatch(
		context.Background(),
		nil,
	)

	require.NoError(t, err)
	assert.Empty(t, ids)
}
