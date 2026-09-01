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

func (persistence *PersistenceService) SaveBatch(
	ctx context.Context,
	links []service.ShortLink,
) error {
	if len(links) == 0 {
		return nil
	}

	tx, err := persistence.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(
			"begin batch transaction: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(
		ctx,
		`
			INSERT INTO short_urls (
				short_url,
				original_url
			)
			VALUES ($1, $2)
			ON CONFLICT (short_url) DO NOTHING
		`,
	)
	if err != nil {
		return fmt.Errorf(
			"prepare batch insert: %w",
			err,
		)
	}
	defer stmt.Close()

	for _, link := range links {
		result, err := stmt.ExecContext(
			ctx,
			link.ID,
			link.OriginalURL,
		)
		if err != nil {
			return fmt.Errorf(
				"insert batch item: %w",
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
				link.ID,
				service.ErrShortLinkIDExists,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"commit batch transaction: %w",
			err,
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
