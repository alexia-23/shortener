package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type deleteRepositoryStub struct {
	started chan []DeleteShortLink
	release chan struct{}
	once    sync.Once
}

func newDeleteRepositoryStub() *deleteRepositoryStub {
	return &deleteRepositoryStub{
		started: make(chan []DeleteShortLink, 1),
		release: make(chan struct{}),
	}
}

func (stub *deleteRepositoryStub) Save(
	_ context.Context,
	_ string,
	_ string,
) error {
	return nil
}

func (stub *deleteRepositoryStub) SaveBatch(
	_ context.Context,
	_ []ShortLink,
) error {
	return nil
}

func (stub *deleteRepositoryStub) Get(
	_ context.Context,
	_ string,
) (string, bool, error) {
	return "", false, nil
}

func (stub *deleteRepositoryStub) DeleteBatch(
	_ context.Context,
	links []DeleteShortLink,
) error {
	copiedLinks := append(
		[]DeleteShortLink(nil),
		links...,
	)

	stub.started <- copiedLinks

	<-stub.release

	return nil
}

func (stub *deleteRepositoryStub) unblock() {
	stub.once.Do(func() {
		close(stub.release)
	})
}

func TestShortLinkService_DeleteUserLinks(t *testing.T) {
	repository := newDeleteRepositoryStub()
	defer repository.unblock()

	shortLinkService := NewShortLinkService(
		repository,
		WithDeleteConfig(
			DeleteConfig{
				FlushInterval: 5 * time.Millisecond,
			},
		),
	)

	t.Cleanup(shortLinkService.Close)

	returned := make(chan struct{})

	go func() {
		shortLinkService.DeleteUserLinks(
			"user-1",
			[]string{
				"first",
				"second",
				"first",
			},
		)

		close(returned)
	}()

	select {
	case <-returned:
		// DeleteUserLinks вернул управление,
		// не дожидаясь фактического удаления.
	case <-time.After(time.Second):
		t.Fatal(
			"DeleteUserLinks must return asynchronously",
		)
	}

	var deletedLinks []DeleteShortLink

	select {
	case deletedLinks = <-repository.started:
	case <-time.After(time.Second):
		t.Fatal(
			"DeleteBatch was not called",
		)
	}

	require.Equal(
		t,
		[]DeleteShortLink{
			{
				ID:     "first",
				UserID: "user-1",
			},
			{
				ID:     "second",
				UserID: "user-1",
			},
		},
		deletedLinks,
	)

	repository.unblock()
}

func (*deleteRepositoryStub) GetByUserID(
	context.Context,
	string,
) ([]ShortLink, error) {
	panic("unexpected GetByUserID call in a deletion test")
}
