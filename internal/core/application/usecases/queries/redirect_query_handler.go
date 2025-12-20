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

	cachedVal, err := h.cache.Get(ctx, q.ShortURL)
	span.AddEvent("retrieval from cache attempt performed")
	if err != nil {
		if !errors.Is(err, errs.ErrObjectNotFound) {
			// Add any other error to span (since not expected)
			span.RecordError(err)
			h.log.Errorw("error getting url from cache", "error", err)
		} else {
			h.log.Warnw("value not found in cache", "short_url", q.ShortURL)
		}
	} else {
		span.AddEvent("value found in cache")
		if cachedVal == "" {
			return RedirectResponse{}, errs.NewObjectNotFoundError("short url", q.ShortURL)
		}

		h.log.Debugw("value found in cache", "short_url", q.ShortURL)
		return RedirectResponse{OriginalURL: cachedVal}, nil
	}

	query := `UPDATE urls SET clicks = clicks + 1 WHERE short_url = $1 RETURNING original_url`
	var originalURL string
	err = h.db.QueryRow(ctx, query, q.ShortURL).Scan(&originalURL)
	span.AddEvent("retrieval from db attempt performed")
	if err != nil {
		span.RecordError(err)
		h.log.Errorw("error getting original url", "error", err)
		if errors.Is(err, pgx.ErrNoRows) {
			// Still cache nil result
			err = h.cache.Set(ctx, q.ShortURL, "")
			span.AddEvent("attempted to save empty value in cache")
			if err != nil {
				span.RecordError(err)
				h.log.Errorw("error saving url to cache", "error", err)
			}

			return RedirectResponse{}, errs.NewObjectNotFoundError("short url", q.ShortURL)
		}

		return RedirectResponse{}, err
	}

	span.AddEvent("url found in db or cache")
	h.log.Debugw("got original url", "original_url", originalURL)

	err = h.cache.Set(ctx, q.ShortURL, originalURL)
	span.AddEvent("attempted to save new value in cache")
	if err != nil {
		span.RecordError(err)
		h.log.Errorw("error saving url to cache", "error", err)
	}

	span.AddEvent("value retrieved and saved to cache successfully")

	return RedirectResponse{
		OriginalURL: originalURL,
	}, nil
}
