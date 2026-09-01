package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alexia-23/shortener/internal/service"
)

type PersistenceService struct {
	db *sql.DB
}

func NewPersistenceService(db *sql.DB) *PersistenceService {
	return &PersistenceService{
		db: db,
	}
}

func (persistence *PersistenceService) PingContext(
	ctx context.Context,
) error {
	return persistence.db.PingContext(ctx)
}

func (persistence *PersistenceService) Save(
	ctx context.Context,
	id string,
	originalURL string,
) error {
	result, err := persistence.db.ExecContext(
		ctx,
		`
			INSERT INTO short_urls (
				short_url,
				original_url
			)
			VALUES ($1, $2)
			ON CONFLICT (short_url) DO NOTHING
		`,
		id,
		originalURL,
	)
	if err != nil {
		return fmt.Errorf(
			"insert short link: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"short link with ID %q: %w",
			id,
			service.ErrShortLinkIDExists,
		)
	}

	return nil
}

func (persistence *PersistenceService) Get(
	ctx context.Context,
	id string,
) (string, bool, error) {
	var originalURL string

	err := persistence.db.QueryRowContext(
		ctx,
		`
			SELECT original_url
			FROM short_urls
			WHERE short_url = $1
		`,
		id,
	).Scan(&originalURL)

	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf(
			"get short link: %w",
			err,
		)
	}

	return originalURL, true, nil
}
