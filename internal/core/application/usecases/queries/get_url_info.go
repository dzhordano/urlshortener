package queries

import (
	"context"
	"errors"
	"time"

	"github.com/dzhordano/urlshortener/internal/core/domain/model"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/dzhordano/urlshortener/internal/pkg/logger"
	"github.com/dzhordano/urlshortener/internal/pkg/tracing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetURLInfoQuery struct {
	ShortURL string
}

func NewGetURLInfoQuery(shortURL string) (GetURLInfoQuery, error) {
	if shortURL == "" {
		return GetURLInfoQuery{}, errs.NewValueIsInvalidError("shortURL")
	}

	return GetURLInfoQuery{
		ShortURL: shortURL,
	}, nil
}

type GetURLInfoResponse struct {
	ID            string
	OriginalURL   string
	ShortURL      string
	Clicks        int
	CreatedAtUTC  time.Time
	ValidUntilUTC time.Time
}

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

	// Get full url info using short url
	query := `
	SELECT id, original_url, short_url, clicks, created_at, valid_until 
	FROM urls
	WHERE short_url = $1`
	var url model.ShortenedURL
	err := h.db.QueryRow(ctx, query, q.ShortURL).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortURL,
		&url.Clicks,
		&url.CreatedAtUTC,
		&url.ValidUntilUTC,
	)
	span.AddEvent("url query db attempt performed")
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, pgx.ErrNoRows) {
			return GetURLInfoResponse{}, errs.NewObjectNotFoundError("short url", q.ShortURL)
		}

		h.log.Error("error getting url info", "error", err)
		return GetURLInfoResponse{}, err
	}

	span.AddEvent("url query db succeeded")
	h.log.Debug("url info", "url", url)

	return GetURLInfoResponse{
		ID:            url.ID.String(),
		OriginalURL:   url.OriginalURL,
		ShortURL:      url.ShortURL,
		Clicks:        url.Clicks,
		CreatedAtUTC:  url.CreatedAtUTC,
		ValidUntilUTC: url.ValidUntilUTC,
	}, nil
}
