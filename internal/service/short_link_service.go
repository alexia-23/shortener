package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

const maxGenerateAttempts = 10

var (
	ErrGenerateUniqueID  = errors.New("failed to generate unique short link ID")
	ErrShortLinkIDExists = errors.New("short link ID already exists")
)

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

type ShortLink struct {
	ID          string
	OriginalURL string
}

type ShortLinkService struct {
	repository ShortLinkRepository
}

func NewShortLinkService(
	repository ShortLinkRepository,
) *ShortLinkService {
	return &ShortLinkService{
		repository: repository,
	}
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
		ids := make([]string, 0, len(originalURLs))
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
			ids = append(ids, id)
			links = append(links, ShortLink{
				ID:          id,
				OriginalURL: originalURL,
			})
		}

		if links == nil {
			continue
		}

		err := service.repository.SaveBatch(ctx, links)
		if err == nil {
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
