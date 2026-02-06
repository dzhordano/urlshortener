package queries

import (
	"context"
	"errors"

	"github.com/dzhordano/urlshortener/internal/core/ports"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/dzhordano/urlshortener/internal/pkg/logger"
	"github.com/dzhordano/urlshortener/internal/pkg/tracing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RedirectQuery struct {
	ShortURL string
}

func NewRedirectQuery(shortURL string) (RedirectQuery, error) {
	if shortURL == "" {
		return RedirectQuery{}, errs.NewValueIsInvalidError("shortURL")
	}

	return RedirectQuery{
		ShortURL: shortURL,
	}, nil
}

type RedirectResponse struct {
	OriginalURL string
}

type RedirectQueryHandler interface {
	Handle(context.Context, RedirectQuery) (RedirectResponse, error)
}

type redirectQueryHandler struct {
	log   logger.Logger
	cache ports.URLCache
	db    *pgxpool.Pool
}

func NewRedirectQueryHandler(
	log logger.Logger,
	cache ports.URLCache,
	db *pgxpool.Pool,
) (RedirectQueryHandler, error) {
	if log == nil {
		return nil, errs.NewValueIsRequiredError("log")
	}

	if cache == nil {
		return nil, errs.NewValueIsRequiredError("cache")
	}

	if db == nil {
		return nil, errs.NewValueIsRequiredError("db")
	}

	return &redirectQueryHandler{
		log:   log,
		cache: cache,
		db:    db,
	}, nil
}

func (h *redirectQueryHandler) Handle(
	ctx context.Context,
	q RedirectQuery,
) (RedirectResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "GetURLInfoQueryHandler.Handle")
	defer span.End()

	// Get shortened url from cache and if found - increment clicks in db.
	// Otherwise, just log cache miss.
	cachedVal, err := h.cache.Get(ctx, q.ShortURL)
	span.AddEvent("retrieval from cache attempt performed")

	// Pretty fried nesting.
	// Basically:
	// Not found? -> log cache miss
	// Any other error? -> log error
	// No error and value is "" (caching absence of value)? -> return ""
	// No error and valud is NOT ""? -> increment click and return it.
	switch {
	case err != nil && errors.Is(err, errs.ErrObjectNotFound):
		h.log.Warn("value not found in cache", "short_url", q.ShortURL)
	case err != nil:
		span.RecordError(err)
		h.log.Error("error getting url from cache", "error", err)
	default:
		if cachedVal == "" {
			return RedirectResponse{}, errs.NewObjectNotFoundError("short url", q.ShortURL)
		}

		// Increment url clicks here.
		query := `
		UPDATE urls
		SET clicks = clicks + 1
		WHERE short_url = $1`
		_, err = h.db.Exec(ctx, query, q.ShortURL)
		if err != nil {
			// record since it's unexpected to happen
			span.RecordError(err)
			h.log.Warn("failed to increment clicks", "short_url", q.ShortURL)
		}

		h.log.Debug("value found in cache", "short_url", q.ShortURL)
		return RedirectResponse{OriginalURL: cachedVal}, nil
	}

	// Update value if url's still valid and return it's original url.
	query := `
	UPDATE urls 
	SET clicks = clicks + 1
	WHERE short_url = $1 AND valid_until > NOW()
	RETURNING original_url`

	var originalURL string
	err = h.db.QueryRow(ctx, query, q.ShortURL).Scan(&originalURL)
	span.AddEvent("retrieval from db attempt performed")
	if err != nil {
		span.RecordError(err)
		h.log.Error("error getting original url", "error", err)
		if errors.Is(err, pgx.ErrNoRows) {
			// Still cache nil result
			err = h.cache.Set(ctx, q.ShortURL, "")
			span.AddEvent("attempted to save empty value in cache")
			if err != nil {
				span.RecordError(err)
				h.log.Error("error saving url to cache", "error", err)
			}

			return RedirectResponse{}, errs.NewObjectNotFoundError("short url", q.ShortURL)
		}

		return RedirectResponse{}, err
	}

	span.AddEvent("url found in db or cache")
	h.log.Debug("got original url", "original_url", originalURL)

	// Cache value for faster next retrieval.
	err = h.cache.Set(ctx, q.ShortURL, originalURL)
	span.AddEvent("attempted to save new value in cache")
	if err != nil {
		span.RecordError(err)
		h.log.Error("error saving url to cache", "error", err)
	}

	span.AddEvent("value retrieved and saved to cache successfully")

	return RedirectResponse{
		OriginalURL: originalURL,
	}, nil
}
