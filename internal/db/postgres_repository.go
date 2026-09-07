package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/alexia-23/shortener/internal/service"
	"github.com/alexia-23/shortener/migrations"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const originalURLUniqueIndexName = "short_urls_original_url_idx"

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	if err := migrations.Up(db); err != nil {
		return nil, fmt.Errorf(
			"apply database migrations: %w",
			err,
		)
	}

	return &PostgresRepository{
		db: db,
	}, nil
}

func (repository *PostgresRepository) PingContext(
	ctx context.Context,
) error {
	return repository.db.PingContext(ctx)
}

func (repository *PostgresRepository) Save(
	ctx context.Context,
	id string,
	originalURL string,
) error {
	_, err := repository.db.ExecContext(
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

	constraintName, uniqueViolation := getUniqueViolationConstraint(err)
	if !uniqueViolation {
		return fmt.Errorf(
			"insert short link: %w",
			err,
		)
	}

	if constraintName == originalURLUniqueIndexName {
		existingID, findErr := repository.findIDByOriginalURL(
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
			"original URL conflict: %w",
			err,
		)
	}

	return fmt.Errorf(
		"short link with ID %q: %w",
		id,
		service.ErrShortLinkIDExists,
	)
}

func (repository *PostgresRepository) findIDByOriginalURL(
	ctx context.Context,
	originalURL string,
) (string, error) {
	var id string

	err := repository.db.QueryRowContext(
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

func getUniqueViolationConstraint(err error) (string, bool) {
	var postgresError *pgconn.PgError

	if !errors.As(err, &postgresError) ||
		postgresError.Code != pgerrcode.UniqueViolation {
		return "", false
	}

	return postgresError.ConstraintName, true
}

func isUniqueViolation(err error) bool {
	_, uniqueViolation := getUniqueViolationConstraint(err)

	return uniqueViolation
}

func (repository *PostgresRepository) SaveBatch(
	ctx context.Context,
	links []service.ShortLink,
) error {
	if len(links) == 0 {
		return nil
	}

	uniqueLinks := make([]service.ShortLink, 0, len(links))
	seenOriginalURLs := make(map[string]struct{}, len(links))

	for _, link := range links {
		if _, exists := seenOriginalURLs[link.OriginalURL]; exists {
			continue
		}

		seenOriginalURLs[link.OriginalURL] = struct{}{}
		uniqueLinks = append(uniqueLinks, link)
	}

	values := make([]string, 0, len(uniqueLinks))
	args := make([]any, 0, len(uniqueLinks)*2)

	for index, link := range uniqueLinks {
		firstPlaceholder := index*2 + 1
		secondPlaceholder := firstPlaceholder + 1

		values = append(
			values,
			fmt.Sprintf(
				"($%d, $%d)",
				firstPlaceholder,
				secondPlaceholder,
			),
		)

		args = append(
			args,
			link.ID,
			link.OriginalURL,
		)
	}

	query := fmt.Sprintf(
		`
			INSERT INTO short_urls (
				short_url,
				original_url
			)
			VALUES %s
			ON CONFLICT (original_url)
			DO UPDATE SET original_url = EXCLUDED.original_url
			RETURNING short_url, original_url
		`,
		strings.Join(values, ", "),
	)

	rows, err := repository.db.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf(
				"short link ID conflict: %w",
				service.ErrShortLinkIDExists,
			)
		}

		return fmt.Errorf(
			"insert short links batch: %w",
			err,
		)
	}
	defer rows.Close()

	idsByOriginalURL := make(
		map[string]string,
		len(uniqueLinks),
	)

	for rows.Next() {
		var id string
		var originalURL string

		if err := rows.Scan(
			&id,
			&originalURL,
		); err != nil {
			return fmt.Errorf(
				"scan batch result: %w",
				err,
			)
		}

		idsByOriginalURL[originalURL] = id
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf(
			"read batch result: %w",
			err,
		)
	}

	for index := range links {
		id, exists := idsByOriginalURL[links[index].OriginalURL]
		if !exists {
			return fmt.Errorf(
				"short link ID not found for original URL %q",
				links[index].OriginalURL,
			)
		}

		links[index].ID = id
	}

	return nil
}

func (repository *PostgresRepository) Get(
	ctx context.Context,
	id string,
) (string, bool, error) {
	var originalURL string

	err := repository.db.QueryRowContext(
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
