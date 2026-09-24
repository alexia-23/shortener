package repository

import (
	"context"
	"testing"

	"github.com/alexia-23/shortener/internal/auth"
	"github.com/alexia-23/shortener/internal/service"
	"github.com/stretchr/testify/require"
)

func TestMemoryRepository_DeleteBatch_OnlyOwnerCanDelete(t *testing.T) {
	repository := NewMemoryRepository()

	ownerCtx := auth.ContextWithUserID(
		context.Background(),
		"owner-user",
	)

	otherUserCtx := auth.ContextWithUserID(
		context.Background(),
		"other-user",
	)

	err := repository.Save(
		ownerCtx,
		"abc123",
		"https://example.com",
	)
	require.NoError(t, err)

	// Чужой пользователь пытается удалить ссылку.
	err = repository.DeleteBatch(
		otherUserCtx,
		[]service.DeleteShortLink{
			{
				ID:     "abc123",
				UserID: "other-user",
			},
		},
	)
	require.NoError(t, err)

	// Ссылка должна остаться доступной.
	originalURL, found, err := repository.Get(
		context.Background(),
		"abc123",
	)
	require.NoError(t, err)

	require.True(t, found)
	require.Equal(
		t,
		"https://example.com",
		originalURL,
	)

	// Теперь ссылку удаляет настоящий владелец.
	err = repository.DeleteBatch(
		ownerCtx,
		[]service.DeleteShortLink{
			{
				ID:     "abc123",
				UserID: "owner-user",
			},
		},
	)
	require.NoError(t, err)

	// Запись существует, но помечена удалённой.
	_, found, err = repository.Get(
		context.Background(),
		"abc123",
	)

	require.True(t, found)
	require.ErrorIs(
		t,
		err,
		service.ErrShortLinkDeleted,
	)
}
