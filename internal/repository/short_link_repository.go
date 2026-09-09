package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/alexia-23/shortener/internal/service"
)

type FileRepository struct {
	memory          *MemoryRepository
	fileStoragePath string
	nextUUID        int
	mutex           sync.RWMutex
}

type storedURL struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFileRepository(
	fileStoragePath string,
) (*FileRepository, error) {
	repository := &FileRepository{
		memory:          NewMemoryRepository(),
		fileStoragePath: fileStoragePath,
		nextUUID:        1,
	}

	if err := repository.loadFromFile(); err != nil {
		return nil, err
	}

	return repository, nil
}

func NewShortLinkRepository(
	fileStoragePath string,
) (*FileRepository, error) {
	return NewFileRepository(fileStoragePath)
}

func (repository *FileRepository) Save(
	ctx context.Context,
	id string,
	originalURL string,
) error {
	return repository.SaveBatch(
		ctx,
		[]service.ShortLink{
			{
				ID:          id,
				OriginalURL: originalURL,
			},
		},
	)
}

func (repository *FileRepository) SaveBatch(
	ctx context.Context,
	links []service.ShortLink,
) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if err := repository.memory.SaveBatch(
		ctx,
		links,
	); err != nil {
		return err
	}

	records := make([]storedURL, 0, len(links))
	for index, link := range links {
		records = append(records, storedURL{
			UUID:        strconv.Itoa(repository.nextUUID + index),
			ShortURL:    link.ID,
			OriginalURL: link.OriginalURL,
		})
	}

	if err := repository.appendBatchToFile(records); err != nil {
		repository.memory.deleteBatch(links)
		return err
	}

	repository.nextUUID += len(records)

	return nil
}

func (repository *FileRepository) Get(
	ctx context.Context,
	id string,
) (string, bool, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	return repository.memory.Get(
		ctx,
		id,
	)
}

func (repository *FileRepository) appendBatchToFile(
	records []storedURL,
) (err error) {
	if len(records) == 0 {
		return nil
	}

	file, err := os.OpenFile(
		repository.fileStoragePath,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0o666,
	)
	if err != nil {
		return fmt.Errorf("open storage file: %w", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("get storage file info: %w", err)
	}

	originalSize := fileInfo.Size()

	defer func() {
		if err == nil {
			return
		}

		if truncateErr := os.Truncate(
			repository.fileStoragePath,
			originalSize,
		); truncateErr != nil {
			err = errors.Join(
				err,
				fmt.Errorf(
					"rollback storage file: %w",
					truncateErr,
				),
			)
		}
	}()

	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close storage file: %w", closeErr)
		}
	}()

	encoder := json.NewEncoder(file)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("encode stored URL: %w", err)
		}
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync storage file: %w", err)
	}

	return nil
}

func (repository *FileRepository) loadFromFile() (err error) {
	file, err := os.Open(repository.fileStoragePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("open storage file: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close storage file: %w", closeErr)
		}
	}()

	decoder := json.NewDecoder(file)

	for {
		var record storedURL

		decodeErr := decoder.Decode(&record)
		if errors.Is(decodeErr, io.EOF) {
			break
		}
		if decodeErr != nil {
			return fmt.Errorf("decode stored URL: %w", decodeErr)
		}

		uuid, conversionErr := strconv.Atoi(record.UUID)
		if conversionErr != nil || uuid < 1 {
			return fmt.Errorf("invalid UUID %q", record.UUID)
		}

		if err := repository.memory.Save(
			context.Background(),
			record.ShortURL,
			record.OriginalURL,
		); err != nil {
			return fmt.Errorf(
				"load short link %q: %w",
				record.ShortURL,
				err,
			)
		}

		if uuid >= repository.nextUUID {
			repository.nextUUID = uuid + 1
		}
	}

	return nil
}
