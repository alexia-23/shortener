package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alexia-23/shortener/internal/service"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	_, err := persistence.db.ExecContext(
		ctx,
		`
			INSERT INTO short_urls (
				short_url,
				original_url
			)
			VALUES ($1, $2)
		`,
		id,
		originalURL,
	)
	if err == nil {
		return nil
	}

	if !isUniqueViolation(err) {
		return fmt.Errorf(
			"insert short link: %w",
			err,
		)
	}

	existingID, findErr := persistence.findIDByOriginalURL(
		ctx,
		originalURL,
	)
	if findErr == nil {
		return &service.OriginalURLExistsError{
			ID: existingID,
		}
	}

	if !errors.Is(findErr, sql.ErrNoRows) {
		return fmt.Errorf(
			"find existing short link after insert conflict: %w",
			findErr,
		)
	}

	return fmt.Errorf(
		"short link with ID %q: %w",
		id,
		service.ErrShortLinkIDExists,
	)
}

func (persistence *PersistenceService) findIDByOriginalURL(
	ctx context.Context,
	originalURL string,
) (string, error) {
	var id string

	err := persistence.db.QueryRowContext(
		ctx,
		`
			SELECT short_url
			FROM short_urls
			WHERE original_url = $1
		`,
		originalURL,
	).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError

	return errors.As(err, &postgresError) &&
		postgresError.Code == pgerrcode.UniqueViolation
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
