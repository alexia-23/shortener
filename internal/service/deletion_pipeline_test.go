package service

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

type recordingDeletionRepository struct {
	mu             sync.Mutex
	batches        [][]DeleteShortLink
	calls          chan []DeleteShortLink
	queryTimedOut  chan struct{}
	waitForTimeout bool
	userLinks      []ShortLink
	userLinksErr   error
}

func (*recordingDeletionRepository) Save(
	context.Context,
	string,
	string,
) error {
	return nil
}

func (*recordingDeletionRepository) SaveBatch(
	context.Context,
	[]ShortLink,
) error {
	return nil
}

func (*recordingDeletionRepository) Get(
	context.Context,
	string,
) (string, bool, error) {
	return "", false, nil
}

func (repo *recordingDeletionRepository) GetByUserID(
	context.Context,
	string,
) ([]ShortLink, error) {
	return append(
		[]ShortLink(nil),
		repo.userLinks...,
	), repo.userLinksErr
}

func (repo *recordingDeletionRepository) DeleteBatch(
	ctx context.Context,
	links []DeleteShortLink,
) error {
	copied := append(
		[]DeleteShortLink(nil),
		links...,
	)

	repo.mu.Lock()
	repo.batches = append(repo.batches, copied)
	repo.mu.Unlock()

	if repo.calls != nil {
		repo.calls <- copied
	}

	if repo.waitForTimeout {
		<-ctx.Done()
		close(repo.queryTimedOut)

		return ctx.Err()
	}

	return nil
}

func (repo *recordingDeletionRepository) snapshot() [][]DeleteShortLink {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	result := make(
		[][]DeleteShortLink,
		len(repo.batches),
	)

	for i := range repo.batches {
		result[i] = append(
			[]DeleteShortLink(nil),
			repo.batches[i]...,
		)
	}

	return result
}

func TestDeleteConfigDefaultsAndOverrides(t *testing.T) {
	defaults := DefaultDeleteConfig()

	if defaults.FlushInterval != 500*time.Millisecond ||
		defaults.MaxWorkers != 8 {
		t.Fatalf(
			"unexpected defaults: %+v",
			defaults,
		)
	}

	repo := &recordingDeletionRepository{}

	service := NewShortLinkService(
		repo,
		WithDeleteConfig(
			DeleteConfig{
				BatchSize:     3,
				FlushInterval: 20 * time.Millisecond,
				MaxWorkers:    2,
			},
		),
	)

	t.Cleanup(service.Close)

	if service.deleteConfig.BatchSize != 3 ||
		service.deleteConfig.FlushInterval != 20*time.Millisecond ||
		service.deleteConfig.MaxWorkers != 2 ||
		service.deleteConfig.StreamBuffer != defaults.StreamBuffer {
		t.Fatalf(
			"unexpected configuration: %+v",
			service.deleteConfig,
		)
	}
}

func TestDeleteConfigRejectsNegativeSettings(t *testing.T) {
	tests := []DeleteConfig{
		{
			BatchSize: -1,
		},
		{
			FlushInterval: -1,
		},
		{
			MaxWorkers: -1,
		},
		{
			StreamBuffer: -1,
		},
		{
			QueryTimeout: -1,
		},
	}

	for _, config := range tests {
		t.Run(
			fmt.Sprintf("%+v", config),
			func(t *testing.T) {
				defer func() {
					if recover() == nil {
						t.Error(
							"constructor must reject a negative deletion setting",
						)
					}
				}()

				NewShortLinkService(
					&recordingDeletionRepository{},
					WithDeleteConfig(config),
				)
			},
		)
	}
}

func TestDeletePipelineFlushesOnBatchSize(t *testing.T) {
	repo := &recordingDeletionRepository{
		calls: make(chan []DeleteShortLink, 4),
	}

	service := NewShortLinkService(
		repo,
		WithDeleteConfig(
			DeleteConfig{
				BatchSize:     2,
				FlushInterval: time.Hour,
			},
		),
	)

	t.Cleanup(service.Close)

	service.DeleteUserLinks(
		"user",
		[]string{
			"first",
			"second",
		},
	)

	select {
	case batch := <-repo.calls:
		if len(batch) != 2 {
			t.Fatalf(
				"batch size = %d, want 2",
				len(batch),
			)
		}

	case <-time.After(2 * time.Second):
		t.Fatal(
			"a full batch must flush without waiting for the timer",
		)
	}
}

func TestDeletePipelineFlushesOnInterval(t *testing.T) {
	repo := &recordingDeletionRepository{
		calls: make(chan []DeleteShortLink, 4),
	}

	service := NewShortLinkService(
		repo,
		WithDeleteConfig(
			DeleteConfig{
				BatchSize:     100,
				FlushInterval: 5 * time.Millisecond,
			},
		),
	)

	t.Cleanup(service.Close)

	service.DeleteUserLinks(
		"user",
		[]string{"first"},
	)

	select {
	case batch := <-repo.calls:
		if len(batch) != 1 ||
			batch[0] != (DeleteShortLink{
				ID:     "first",
				UserID: "user",
			}) {
			t.Fatalf(
				"unexpected batch: %+v",
				batch,
			)
		}

	case <-time.After(2 * time.Second):
		t.Fatal(
			"a partial batch must flush on the interval",
		)
	}
}

func TestCloseFlushesPartialBatchAndDeduplicatesRequest(
	t *testing.T,
) {
	repo := &recordingDeletionRepository{}

	service := NewShortLinkService(
		repo,
		WithDeleteConfig(
			DeleteConfig{
				FlushInterval: time.Hour,
			},
		),
	)

	service.DeleteUserLinks(
		"user",
		[]string{
			"",
			"first",
			"first",
			"second",
			"",
		},
	)

	service.Close()
	service.Close()

	got := repo.snapshot()

	if len(got) != 1 ||
		len(got[0]) != 2 ||
		got[0][0].ID != "first" ||
		got[0][1].ID != "second" {
		t.Fatalf(
			"unexpected deduplicated batch: %+v",
			got,
		)
	}

	for _, link := range got[0] {
		if link.UserID != "user" {
			t.Fatalf(
				"owner not preserved: %+v",
				link,
			)
		}
	}
}

func TestCloseWithoutDeletionAndEmptyRequests(t *testing.T) {
	repo := &recordingDeletionRepository{}
	service := NewShortLinkService(repo)

	service.DeleteUserLinks(
		"",
		[]string{"first"},
	)

	service.DeleteUserLinks(
		"user",
		nil,
	)

	service.DeleteUserLinks(
		"user",
		[]string{"", ""},
	)

	service.Close()
	service.Close()

	service.DeleteUserLinks(
		"user",
		[]string{"after-close"},
	)

	if len(repo.snapshot()) != 0 {
		t.Fatal(
			"empty requests or calls after Close must not start a deletion worker",
		)
	}
}

func TestDeletePipelineCombinesRequestsAndPreservesOwners(
	t *testing.T,
) {
	repo := &recordingDeletionRepository{}

	service := NewShortLinkService(
		repo,
		WithDeleteConfig(
			DeleteConfig{
				BatchSize:     100,
				FlushInterval: time.Hour,
			},
		),
	)

	service.DeleteUserLinks(
		"alice",
		[]string{
			"same-id",
			"alice-only",
		},
	)

	service.DeleteUserLinks(
		"bob",
		[]string{
			"same-id",
			"bob-only",
		},
	)

	service.Close()

	batches := repo.snapshot()

	if len(batches) != 1 ||
		len(batches[0]) != 4 {
		t.Fatalf(
			"requests were not combined in one batch: %+v",
			batches,
		)
	}

	seen := make(map[DeleteShortLink]bool)

	for _, link := range batches[0] {
		seen[link] = true
	}

	for _, want := range []DeleteShortLink{
		{
			ID:     "same-id",
			UserID: "alice",
		},
		{
			ID:     "alice-only",
			UserID: "alice",
		},
		{
			ID:     "same-id",
			UserID: "bob",
		},
		{
			ID:     "bob-only",
			UserID: "bob",
		},
	} {
		if !seen[want] {
			t.Errorf(
				"missing deletion with original owner: %+v",
				want,
			)
		}
	}
}

func TestDeletePipelineConcurrentEnqueueAndDrain(t *testing.T) {
	repo := &recordingDeletionRepository{}

	service := NewShortLinkService(
		repo,
		WithDeleteConfig(
			DeleteConfig{
				BatchSize:     7,
				FlushInterval: time.Hour,
				MaxWorkers:    3,
				StreamBuffer:  2,
			},
		),
	)

	const requests = 100

	var wg sync.WaitGroup

	for i := 0; i < requests; i++ {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			service.DeleteUserLinks(
				fmt.Sprintf(
					"user-%d",
					index,
				),
				[]string{
					fmt.Sprintf(
						"id-%d",
						index,
					),
				},
			)
		}(i)
	}

	wg.Wait()
	service.Close()

	seen := make(map[DeleteShortLink]bool)

	for _, batch := range repo.snapshot() {
		if len(batch) > 7 {
			t.Fatalf(
				"batch exceeded configured size: %d",
				len(batch),
			)
		}

		for _, link := range batch {
			if seen[link] {
				t.Fatalf(
					"duplicate deletion: %+v",
					link,
				)
			}

			seen[link] = true
		}
	}

	if len(seen) != requests {
		t.Fatalf(
			"drained %d requests, want %d",
			len(seen),
			requests,
		)
	}
}

func TestDeletePipelineConcurrentCloseAndEnqueue(t *testing.T) {
	service := NewShortLinkService(
		&recordingDeletionRepository{},
		WithDeleteConfig(
			DeleteConfig{
				BatchSize:     3,
				FlushInterval: time.Millisecond,
				StreamBuffer:  1,
				MaxWorkers:    2,
			},
		),
	)

	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			if index%4 == 0 {
				service.Close()
			} else {
				service.DeleteUserLinks(
					"user",
					[]string{
						fmt.Sprint(index),
					},
				)
			}
		}(i)
	}

	wg.Wait()
	service.Close()
}

func TestFanInBoundsWriterGoroutines(t *testing.T) {
	const requests = 200

	config := DefaultDeleteConfig()
	config.MaxWorkers = 3
	config.BatchSize = 1

	streams := make(
		chan (<-chan DeleteShortLink),
		requests,
	)

	for i := 0; i < requests; i++ {
		stream := make(
			chan DeleteShortLink,
			1,
		)

		stream <- DeleteShortLink{
			ID:     fmt.Sprint(i),
			UserID: "user",
		}

		close(stream)
		streams <- stream
	}

	close(streams)

	before := runtime.NumGoroutine()

	output := fanIn(
		streams,
		config,
	)

	// Let output fill without reading:
	// writers cannot complete past this point.
	deadline := time.Now().Add(time.Second)

	for len(output) == 0 &&
		time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}

	time.Sleep(20 * time.Millisecond)

	after := runtime.NumGoroutine()

	// Three writers plus one coordinator;
	// allow noise from the test runtime.
	if after-before > config.MaxWorkers+5 {
		t.Errorf(
			"goroutine growth = %d for %d streams, want a bound near %d",
			after-before,
			requests,
			config.MaxWorkers,
		)
	}

	count := 0

	timeout := time.NewTimer(
		2 * time.Second,
	)
	defer timeout.Stop()

	for {
		select {
		case _, ok := <-output:
			if !ok {
				if count != requests {
					t.Fatalf(
						"fan-in forwarded %d records, want %d",
						count,
						requests,
					)
				}

				return
			}

			count++

		case <-timeout.C:
			t.Fatal(
				"fan-in did not close after all streams were drained",
			)
		}
	}
}

func TestDeletionQueryHasTimeout(t *testing.T) {
	repo := &recordingDeletionRepository{
		waitForTimeout: true,
		queryTimedOut:  make(chan struct{}),
	}

	service := NewShortLinkService(
		repo,
		WithDeleteConfig(
			DeleteConfig{
				BatchSize:    1,
				QueryTimeout: 5 * time.Millisecond,
			},
		),
	)

	t.Cleanup(service.Close)

	service.DeleteUserLinks(
		"user",
		[]string{"id"},
	)

	select {
	case <-repo.queryTimedOut:

	case <-time.After(2 * time.Second):
		t.Fatal(
			"DeleteBatch context must expire",
		)
	}
}

func TestGetUserLinksUsesTypedRepositoryAndSorts(
	t *testing.T,
) {
	repo := &recordingDeletionRepository{
		userLinks: []ShortLink{
			{
				ID: "z",
			},
			{
				ID: "a",
			},
		},
	}

	service := NewShortLinkService(repo)
	t.Cleanup(service.Close)

	links, err := service.GetUserLinks(
		context.Background(),
		"user",
	)

	if err != nil ||
		len(links) != 2 ||
		links[0].ID != "a" ||
		links[1].ID != "z" {
		t.Fatalf(
			"GetUserLinks() = %+v, %v; want sorted links",
			links,
			err,
		)
	}

	want := errors.New("lookup failed")
	repo.userLinksErr = want

	if _, err := service.GetUserLinks(
		context.Background(),
		"user",
	); !errors.Is(err, want) {
		t.Fatalf(
			"GetUserLinks() error = %v, want %v",
			err,
			want,
		)
	}
}
