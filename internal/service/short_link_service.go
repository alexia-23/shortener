package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"
)

const maxGenerateAttempts = 10

var (
	ErrGenerateUniqueID  = errors.New("failed to generate unique short link ID")
	ErrShortLinkIDExists = errors.New("short link ID already exists")
	ErrOriginalURLExists = errors.New("original URL already exists")
	ErrShortLinkDeleted  = errors.New("short link is deleted")
)

type OriginalURLExistsError struct {
	ID string
}

func (err *OriginalURLExistsError) Error() string {
	return fmt.Sprintf(
		"original URL already has short link with ID %q",
		err.ID,
	)
}

func (err *OriginalURLExistsError) Unwrap() error {
	return ErrOriginalURLExists
}

type ShortLinkRepository interface {
	Save(
		ctx context.Context,
		id string,
		originalURL string,
	) error

	SaveBatch(
		ctx context.Context,
		links []ShortLink,
	) error

	Get(
		ctx context.Context,
		id string,
	) (string, bool, error)
}

type DeleteShortLinkRepository interface {
	DeleteBatch(
		ctx context.Context,
		links []DeleteShortLink,
	) error
}

type UserLinksRepository interface {
	GetByUserID(
		ctx context.Context,
		userID string,
	) ([]ShortLink, error)
}

// URLRepository is the complete dependency required by ShortLinkService.
// Implementations missing user lookup or deletion fail to compile at wiring.
type URLRepository interface {
	ShortLinkRepository
	UserLinksRepository
	DeleteShortLinkRepository
}

type ShortLink struct {
	ID          string
	OriginalURL string
	UserID      string
}

type DeleteShortLink struct {
	ID     string
	UserID string
}

type ShortLinkService struct {
	repository    URLRepository
	deleteConfig  DeleteConfig
	deleteStreams chan (<-chan DeleteShortLink)
	deleteOnce    sync.Once
	deleteMu      sync.RWMutex
	deleteClosed  bool
	deleteDone    chan struct{}
}

func NewShortLinkService(
	repository URLRepository,
	options ...Option,
) *ShortLinkService {
	if repository == nil {
		panic("service: nil repository")
	}

	service := &ShortLinkService{
		repository:   repository,
		deleteConfig: DefaultDeleteConfig(),
		deleteDone:   make(chan struct{}),
	}

	for _, option := range options {
		option(service)
	}

	service.deleteStreams = make(
		chan (<-chan DeleteShortLink),
		service.deleteConfig.StreamBuffer,
	)

	return service
}

func (service *ShortLinkService) CreateShortLink(
	ctx context.Context,
	originalURL string,
) (string, error) {
	for attempt := 0; attempt < maxGenerateAttempts; attempt++ {
		id, err := generateID()
		if err != nil {
			return "", err
		}

		err = service.repository.Save(
			ctx,
			id,
			originalURL,
		)
		if err == nil {
			return id, nil
		}

		if errors.Is(err, ErrShortLinkIDExists) {
			continue
		}

		return "", fmt.Errorf(
			"save short link: %w",
			err,
		)
	}

	return "", ErrGenerateUniqueID
}

func (service *ShortLinkService) CreateShortLinksBatch(
	ctx context.Context,
	originalURLs []string,
) ([]string, error) {
	if len(originalURLs) == 0 {
		return []string{}, nil
	}

	for attempt := 0; attempt < maxGenerateAttempts; attempt++ {
		links := make([]ShortLink, 0, len(originalURLs))
		generatedIDs := make(map[string]struct{}, len(originalURLs))

		for _, originalURL := range originalURLs {
			id, err := generateID()
			if err != nil {
				return nil, err
			}

			if _, exists := generatedIDs[id]; exists {
				links = nil
				break
			}

			generatedIDs[id] = struct{}{}

			links = append(
				links,
				ShortLink{
					ID:          id,
					OriginalURL: originalURL,
				},
			)
		}

		if links == nil {
			continue
		}

		err := service.repository.SaveBatch(
			ctx,
			links,
		)
		if err == nil {
			ids := make([]string, len(links))

			for index, link := range links {
				ids[index] = link.ID
			}

			return ids, nil
		}

		if errors.Is(err, ErrShortLinkIDExists) {
			continue
		}

		return nil, fmt.Errorf(
			"save short links batch: %w",
			err,
		)
	}

	return nil, ErrGenerateUniqueID
}

func (service *ShortLinkService) GetSourceLink(
	ctx context.Context,
	id string,
) (string, bool, error) {
	return service.repository.Get(
		ctx,
		id,
	)
}

func (service *ShortLinkService) GetUserLinks(
	ctx context.Context,
	userID string,
) ([]ShortLink, error) {
	links, err := service.repository.GetByUserID(
		ctx,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get user links: %w",
			err,
		)
	}

	sort.Slice(
		links,
		func(i, j int) bool {
			return links[i].ID < links[j].ID
		},
	)

	return links, nil
}

// DeleteUserLinks queues a request without waiting for database deletion.
// A full bounded queue applies backpressure instead of spawning more goroutines.
func (service *ShortLinkService) DeleteUserLinks(
	userID string,
	ids []string,
) {
	if userID == "" || len(ids) == 0 {
		return
	}

	stream := make(
		chan DeleteShortLink,
		len(ids),
	)

	seen := make(
		map[string]struct{},
		len(ids),
	)

	for _, id := range ids {
		if id == "" {
			continue
		}

		if _, exists := seen[id]; exists {
			continue
		}

		seen[id] = struct{}{}

		stream <- DeleteShortLink{
			ID:     id,
			UserID: userID,
		}
	}

	close(stream)

	if len(seen) == 0 {
		return
	}

	// Hold a read lock until enqueueing is finished so Close cannot close
	// deleteStreams while a concurrent caller is sending to it.
	service.deleteMu.RLock()
	defer service.deleteMu.RUnlock()

	if service.deleteClosed {
		return
	}

	service.deleteOnce.Do(func() {
		merged := fanIn(
			service.deleteStreams,
			service.deleteConfig,
		)

		go func() {
			defer close(service.deleteDone)

			service.processDeleteBatches(merged)
		}()
	})

	service.deleteStreams <- stream
}

// Close drains all accepted deletion requests, flushes the last batch and
// waits for the background pipeline to stop. It is safe to call more than once.
// The HTTP server must stop accepting requests before calling Close.
func (service *ShortLinkService) Close() {
	service.deleteMu.Lock()

	if !service.deleteClosed {
		service.deleteClosed = true
		close(service.deleteStreams)

		// If deletion was never started, there is no worker to close deleteDone.
		service.deleteOnce.Do(func() {
			close(service.deleteDone)
		})
	}

	service.deleteMu.Unlock()

	<-service.deleteDone
}

func fanIn(
	streams <-chan (<-chan DeleteShortLink),
	config DeleteConfig,
) <-chan DeleteShortLink {
	result := make(
		chan DeleteShortLink,
		config.BatchSize,
	)

	semaphore := make(
		chan struct{},
		config.MaxWorkers,
	)

	go func() {
		var wg sync.WaitGroup

		for stream := range streams {
			// Acquire BEFORE starting a goroutine. Acquiring inside it would
			// still allow an unbounded number of goroutines waiting for a slot.
			semaphore <- struct{}{}

			wg.Add(1)

			go func(input <-chan DeleteShortLink) {
				defer wg.Done()

				defer func() {
					<-semaphore
				}()

				for link := range input {
					result <- link
				}
			}(stream)
		}

		wg.Wait()
		close(result)
	}()

	return result
}

func (service *ShortLinkService) processDeleteBatches(
	input <-chan DeleteShortLink,
) {
	ticker := time.NewTicker(
		service.deleteConfig.FlushInterval,
	)
	defer ticker.Stop()

	batch := make(
		[]DeleteShortLink,
		0,
		service.deleteConfig.BatchSize,
	)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			service.deleteConfig.QueryTimeout,
		)

		err := service.repository.DeleteBatch(
			ctx,
			batch,
		)

		cancel()

		if err != nil {
			slog.Error(
				"failed to delete short links",
				"error",
				err,
				"batch_size",
				len(batch),
			)
		}

		batch = batch[:0]
	}

	for {
		select {
		case link, ok := <-input:
			if !ok {
				flush()
				return
			}

			batch = append(
				batch,
				link,
			)

			if len(batch) >= service.deleteConfig.BatchSize {
				flush()
			}

		case <-ticker.C:
			flush()
		}
	}
}

func generateID() (string, error) {
	randomBytes := make([]byte, 6)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf(
			"generate random ID: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(
		randomBytes,
	), nil
}
