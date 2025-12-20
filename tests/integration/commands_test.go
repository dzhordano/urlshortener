package integration_test

import (
	"context"
	"time"

	"github.com/dzhordano/urlshortener/internal/core/application/usecases/commands"
	"github.com/dzhordano/urlshortener/internal/core/domain/model"
	"go.uber.org/zap/zaptest"
)

func (s *Suite) TestShortenURLCommandHandler_Success() {
	ctx := context.Background()
	req := commands.ShortenURLCommand{
		OriginalURL: "http://example.com",
	}

	handler, err := commands.NewShortenURLCommandHandler(
		zaptest.NewLogger(s.T()).Sugar(),
		s.cache, s.urlRepo,
	)
	s.Require().NoError(err)

	resp, err := handler.Handle(ctx, req)
	s.Require().NoError(err)

	valueFromDB, err := s.urlRepo.GetByOriginalURL(ctx, req.OriginalURL)
	s.Require().NoError(err)

	valueFromCache, err := s.cache.Get(ctx, valueFromDB.ShortURL)
	s.Require().NoError(err)

	// Check if value is in db
	s.Equal(resp, valueFromDB.ShortURL)
	// Check whether original url is saved (lulz)
	s.Equal(req.OriginalURL, valueFromDB.OriginalURL)
	// Check if cache did save shortened url value
	s.Equal(req.OriginalURL, valueFromCache)
	// Check if clicks are saved correctly
	s.Equal(0, valueFromDB.Clicks)
	// Check if time isn't a nil value
	s.NotEmpty(valueFromDB.CreatedAtUTC)
}

func (s *Suite) TestShortenURLCommandHandler_ReturnExisting() {
	ctx := context.Background()
	req := commands.ShortenURLCommand{
		OriginalURL: "http://example.com",
	}

	// Preliminaries
	// Save value to DB
	shortenedURL := &model.ShortenedURL{
		OriginalURL:  "http://example.com",
		ShortURL:     "SOMEURL",
		Clicks:       1,
		CreatedAtUTC: time.Now().UTC(),
	}
	err := s.urlRepo.Save(ctx, shortenedURL)
	s.Require().NoError(err)

	handler, err := commands.NewShortenURLCommandHandler(
		zaptest.NewLogger(s.T()).Sugar(),
		s.cache, s.urlRepo,
	)
	s.Require().NoError(err)

	resp, err := handler.Handle(ctx, req)
	s.Require().NoError(err)

	valueFromDB, err := s.urlRepo.GetByOriginalURL(ctx, req.OriginalURL)
	s.Require().NoError(err)

	// Check if returning value is correct
	s.Equal(resp, shortenedURL.ShortURL)
	// Check if url is correctly saved in DB
	s.Equal(shortenedURL.OriginalURL, valueFromDB.OriginalURL)
	s.Equal(shortenedURL.ShortURL, valueFromDB.ShortURL)
	s.Equal(shortenedURL.Clicks, valueFromDB.Clicks)
	// Seems like there is a minor difference when saving to db
	s.Equal(shortenedURL.CreatedAtUTC.Year(), valueFromDB.CreatedAtUTC.Year())
	s.Equal(shortenedURL.CreatedAtUTC.Day(), valueFromDB.CreatedAtUTC.Day())
	s.Equal(shortenedURL.CreatedAtUTC.Minute(), valueFromDB.CreatedAtUTC.Minute())
	s.Equal(shortenedURL.CreatedAtUTC.Second(), valueFromDB.CreatedAtUTC.Second())

	// Check presence of value in cache, since we much cache after retrieval
	valueFromCache, err := s.cache.Get(ctx, shortenedURL.ShortURL)
	s.Require().NoError(err)

	s.Equal(shortenedURL.OriginalURL, valueFromCache)
}
