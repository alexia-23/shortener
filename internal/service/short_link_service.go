package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const maxGenerateAttempts = 10

var ErrGenerateUniqueID = fmt.Errorf("failed to generate unique short link ID")

type ShortLinkRepository interface {
	Save(id string, originalURL string) bool
	Get(id string) (string, bool)
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
	originalURL string,
) (string, error) {
	for attempt := 0; attempt < maxGenerateAttempts; attempt++ {
		id, err := generateID()
		if err != nil {
			return "", err
		}

		if service.repository.Save(id, originalURL) {
			return id, nil
		}
	}

	return "", ErrGenerateUniqueID
}

func (service *ShortLinkService) GetSourceLink(
	id string,
) (string, bool) {
	return service.repository.Get(id)
}

func generateID() (string, error) {
	randomBytes := make([]byte, 6)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("generate random ID: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}
