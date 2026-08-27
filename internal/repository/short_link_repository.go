package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
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
	id string,
	originalURL string,
) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if err := repository.memory.Save(id, originalURL); err != nil {
		return err
	}

	record := storedURL{
		UUID:        strconv.Itoa(repository.nextUUID),
		ShortURL:    id,
		OriginalURL: originalURL,
	}

	if err := repository.appendToFile(record); err != nil {
		repository.memory.delete(id)
		return err
	}

	repository.nextUUID++

	return nil
}

func (repository *FileRepository) Get(
	id string,
) (string, bool) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	return repository.memory.Get(id)
}

func (repository *FileRepository) appendToFile(
	record storedURL,
) (err error) {
	file, err := os.OpenFile(
		repository.fileStoragePath,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0o666,
	)
	if err != nil {
		return fmt.Errorf("open storage file: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close storage file: %w", closeErr)
		}
	}()

	if err := json.NewEncoder(file).Encode(record); err != nil {
		return fmt.Errorf("encode stored URL: %w", err)
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
