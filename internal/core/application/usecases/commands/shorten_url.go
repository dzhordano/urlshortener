//nolint:nolintlint,exhaustruct
package commands

import (
	"context"

	"github.com/dzhordano/urlshortener/internal/core/domain/model"
	"github.com/dzhordano/urlshortener/internal/core/ports"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/dzhordano/urlshortener/internal/pkg/logger"
	"github.com/dzhordano/urlshortener/internal/pkg/tracing"
)

type ShortenURLCommand struct {
	OriginalURL string
}

func NewShortenURLCommand(url string) (ShortenURLCommand, error) {
	if url == "" {
		return ShortenURLCommand{}, errs.NewValueIsInvalidError("url")
	}

	return ShortenURLCommand{OriginalURL: url}, nil
}

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
		h.log.Error("error creating new shortened url", "error", err)
		return "", err
	}

	span.AddEvent("shortened url created")
	h.log.Debug("shortened url", "short_url", url.ShortURL)

	// If shortened url will contain non-unique short url (collision)
	// this will result in an error (unique constraint) because no retry logic :p.
	err = h.urlRepo.Save(ctx, url)
	span.AddEvent("shortened url save attempt performed")
	if err != nil {
		span.RecordError(err)
		h.log.Error("error saving url", "error", err)
		return "", err
	}

	span.AddEvent("shortened url saved or retrieved from db")
	h.log.Debug("url saved or found in db", "url", url)

	err = h.cache.Set(ctx, url.ShortURL, url.OriginalURL)
	if err != nil {
		span.RecordError(err)
		h.log.Error("error saving url to cache", "error", err)
	}

	span.AddEvent("shortened url saved")
	h.log.Debug("short url saved to cache", "short_url", url.ShortURL)

	return url.ShortURL, nil
}
