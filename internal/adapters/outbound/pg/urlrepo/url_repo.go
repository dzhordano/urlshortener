package urlrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/dzhordano/urlshortener/internal/core/domain/model"
	"github.com/dzhordano/urlshortener/internal/core/ports"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	urlsTable = "urls"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) (ports.URLRepository, error) {
	if db == nil {
		return nil, errs.NewValueIsRequiredError("db")
	}

	return &Repository{
		db: db,
	}, nil
}

func (r *Repository) Save(ctx context.Context, url *model.ShortenedURL) error {
	const op = "UrlRepo.Save"

	query := fmt.Sprintf(`INSERT INTO %s (original_url, short_url, clicks, created_at_utc)
	VALUES ($1, $2, $3, $4)`, urlsTable)

	_, err := r.db.Exec(ctx, query, url.OriginalURL, url.ShortURL, url.Clicks, url.CreatedAtUTC)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf(
				"%s: %w",
				op, errs.NewObjectAlreadyExistsError("originalURL", url.OriginalURL),
			)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) GetByShortenedURL(
	ctx context.Context,
	shortenedURL string,
) (*model.ShortenedURL, error) {
	const op = "UrlRepo.GetByShortenedURL"

	query := fmt.Sprintf(
		`SELECT original_url, short_url, clicks, created_at_utc FROM %s WHERE short_url = $1`,
		urlsTable,
	)

	var url model.ShortenedURL
	err := r.db.QueryRow(ctx, query, shortenedURL).Scan(
		&url.OriginalURL,
		&url.ShortURL,
		&url.Clicks,
		&url.CreatedAtUTC,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"%s: %w",
				op, errs.NewObjectNotFoundError("shortenedURL", shortenedURL),
			)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &url, nil
}

func (r *Repository) GetByOriginalURL(
	ctx context.Context,
	originalURL string,
) (*model.ShortenedURL, error) {
	const op = "UrlRepo.GetByShortenedURL"

	query := fmt.Sprintf(
		`SELECT original_url, short_url, clicks, created_at_utc FROM %s WHERE original_url = $1`,
		urlsTable,
	)

	var url model.ShortenedURL
	err := r.db.QueryRow(ctx, query, originalURL).Scan(
		&url.OriginalURL,
		&url.ShortURL,
		&url.Clicks,
		&url.CreatedAtUTC,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"%s: %w",
				op, errs.NewObjectNotFoundError("originalURL", originalURL),
			)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &url, nil
}
