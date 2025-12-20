package queries

import (
	"context"
	"errors"

	"github.com/dzhordano/urlshortener/internal/core/domain/model"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/dzhordano/urlshortener/internal/pkg/logger"
	"github.com/dzhordano/urlshortener/internal/pkg/tracing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetURLInfoQueryHandler interface {
	Handle(context.Context, GetURLInfoQuery) (GetURLInfoResponse, error)
}

type getURLInfoQueryHandler struct {
	log logger.Logger
	db  *pgxpool.Pool
}

func NewGetURLInfoQueryHandler(
	log logger.Logger,
	db *pgxpool.Pool,
) (GetURLInfoQueryHandler, error) {
	if log == nil {
		return nil, errs.NewValueIsRequiredError("log")
	}

	if db == nil {
		return nil, errs.NewValueIsRequiredError("db")
	}

	return &getURLInfoQueryHandler{
		log: log,
		db:  db,
	}, nil
}

func (h *getURLInfoQueryHandler) Handle(
	ctx context.Context,
	q GetURLInfoQuery,
) (GetURLInfoResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "GetURLInfoQueryHandler.Handle")
	defer span.End()

	query := `SELECT original_url, short_url, clicks, created_at_utc FROM urls WHERE short_url = $1`
	var url model.ShortenedURL
	err := h.db.QueryRow(ctx, query, q.ShortURL).Scan(
		&url.OriginalURL,
		&url.ShortURL,
		&url.Clicks,
		&url.CreatedAtUTC,
	)
	span.AddEvent("url query db attempt performed")
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, pgx.ErrNoRows) {
			return GetURLInfoResponse{}, errs.NewObjectNotFoundError("short url", q.ShortURL)
		}

		h.log.Errorw("error getting url info", "error", err)
		return GetURLInfoResponse{}, err
	}

	span.AddEvent("url query db succeeded")
	h.log.Debugw("url info", "url", url)

	return GetURLInfoResponse{
		OriginalURL:  url.OriginalURL,
		ShortURL:     url.ShortURL,
		Clicks:       url.Clicks,
		CreatedAtUTC: url.CreatedAtUTC,
	}, nil
}
