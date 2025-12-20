//nolint:nolintlint,exhaustruct
package commands

import (
	"context"
	"errors"

	"github.com/dzhordano/urlshortener/internal/core/domain/model"
	"github.com/dzhordano/urlshortener/internal/core/ports"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/dzhordano/urlshortener/internal/pkg/logger"
	"github.com/dzhordano/urlshortener/internal/pkg/tracing"
)

type ShortenURLCommandHandler interface {
	Handle(context.Context, ShortenURLCommand) (string, error)
}

type shortenURLCommandHandler struct {
	log     logger.Logger
	cache   ports.URLCache
	urlRepo ports.URLRepository
}

func NewShortenURLCommandHandler(
	log logger.Logger,
	cache ports.URLCache,
	urlRepo ports.URLRepository,
) (ShortenURLCommandHandler, error) {
	if log == nil {
		return nil, errs.NewValueIsRequiredError("log")
	}

	if cache == nil {
		return nil, errs.NewValueIsRequiredError("cache")
	}

	if urlRepo == nil {
		return nil, errs.NewValueIsRequiredError("urlRepo")
	}

	return &shortenURLCommandHandler{
		log:     log,
		cache:   cache,
		urlRepo: urlRepo,
	}, nil
}

func (h *shortenURLCommandHandler) Handle(
	ctx context.Context,
	cmd ShortenURLCommand,
) (string, error) {
	ctx, span := tracing.StartSpan(ctx, "ShortenURLCommandHandler.Handle")
	defer span.End()

	url, err := model.NewShortenedURL(cmd.OriginalURL)
	if err != nil {
		span.RecordError(err)
		h.log.Errorw("error creating new shortened url", "error", err)
		return "", err
	}

	span.AddEvent("shortened url created")
	h.log.Debugw("shortened url", "short_url", url.ShortURL)

	err = h.urlRepo.Save(ctx, url)
	span.AddEvent("shortened url save attempt performed")
	if err != nil {
		if !errors.Is(err, errs.ErrObjectAlreadyExists) {
			span.RecordError(err)
			h.log.Errorw("error saving url", "error", err)
			return "", err
		}

		span.AddEvent("shortened url already exists")
		h.log.Debugw("url already exists", "url", url)

		url, err = h.urlRepo.GetByOriginalURL(ctx, url.OriginalURL)
		if err != nil {
			span.RecordError(err)
			return "", err
		}
	}

	span.AddEvent("shortened url saved or retrieved from db")
	h.log.Debugw("url saved or found in db", "url", url)

	err = h.cache.Set(ctx, url.ShortURL, url.OriginalURL)
	if err != nil {
		span.RecordError(err)
		h.log.Errorw("error saving url to cache", "error", err)
	}

	span.AddEvent("shortened url saved")
	h.log.Debugw("short url saved to cache", "short_url", url.ShortURL)

	return url.ShortURL, nil
}
