package integration_test

import (
	"context"
	"time"

	"github.com/dzhordano/urlshortener/internal/core/application/usecases/queries"
	"github.com/dzhordano/urlshortener/internal/core/domain/model"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"go.uber.org/zap/zaptest"
)

func (s *Suite) TestGetURLInfoQueryHandler_Found() {
	ctx := context.Background()
	query := queries.GetURLInfoQuery{
		ShortURL: "SOMEURL",
	}

	// Save shortened url before retrieval
	shortenedURL := &model.ShortenedURL{
		OriginalURL:  "http://example.com",
		ShortURL:     "SOMEURL",
		Clicks:       1,
		CreatedAtUTC: time.Now().UTC(),
	}
	err := s.urlRepo.Save(ctx, shortenedURL)
	s.Require().NoError(err)

	handler, err := queries.NewGetURLInfoQueryHandler(
		zaptest.NewLogger(s.T()).Sugar(), s.pgxPool)
	s.Require().NoError(err)

	resp, err := handler.Handle(ctx, query)
	s.Require().NoError(err)

	valueFromDB, err := s.urlRepo.GetByShortenedURL(ctx, query.ShortURL)
	s.Require().NoError(err)

	// idk why but why not ig
	s.Equal(resp.ShortURL, query.ShortURL)
	// Check returned values with values in DB
	s.Equal(resp.OriginalURL, valueFromDB.OriginalURL)
	s.Equal(resp.ShortURL, valueFromDB.ShortURL)
	s.Equal(resp.Clicks, valueFromDB.Clicks)
	s.Equal(resp.CreatedAtUTC.Year(), valueFromDB.CreatedAtUTC.Year())
	s.Equal(resp.CreatedAtUTC.Day(), valueFromDB.CreatedAtUTC.Day())
	s.Equal(resp.CreatedAtUTC.Minute(), valueFromDB.CreatedAtUTC.Minute())
	s.Equal(resp.CreatedAtUTC.Second(), valueFromDB.CreatedAtUTC.Second())

	// No need to check cache since no cache is used in this method
}

func (s *Suite) TestGetURLInfoQueryHandler_NotFound() {
	ctx := context.Background()
	query := queries.GetURLInfoQuery{
		ShortURL: "SOMEURL",
	}

	handler, err := queries.NewGetURLInfoQueryHandler(
		zaptest.NewLogger(s.T()).Sugar(), s.pgxPool)
	s.Require().NoError(err)

	resp, handlerErr := handler.Handle(ctx, query)

	valueFromDB, repoErr := s.urlRepo.GetByShortenedURL(ctx, query.ShortURL)

	// Kind of depends on specific error type used in both methods :/
	s.Require().ErrorIs(repoErr, errs.ErrObjectNotFound)
	s.Require().ErrorIs(handlerErr, errs.ErrObjectNotFound)
	s.Empty(resp)
	s.Nil(valueFromDB)
}

func (s *Suite) TestRedirect_Found() {
	ctx := context.Background()
	query := queries.RedirectQuery{
		ShortURL: "SOMEURL",
	}

	shortenedURL := &model.ShortenedURL{
		OriginalURL:  "http://example.com",
		ShortURL:     "SOMEURL",
		Clicks:       1,
		CreatedAtUTC: time.Now().UTC(),
	}
	err := s.urlRepo.Save(ctx, shortenedURL)
	s.Require().NoError(err)

	handler, err := queries.NewRedirectQueryHandler(
		zaptest.NewLogger(s.T()).Sugar(), s.cache, s.pgxPool)
	s.Require().NoError(err)

	resp, err := handler.Handle(ctx, query)
	s.Require().NoError(err)

	valueFromDB, err := s.urlRepo.GetByShortenedURL(ctx, query.ShortURL)
	s.Require().NoError(err)

	valueFromCache, err := s.cache.Get(ctx, query.ShortURL)
	s.Require().NoError(err)

	// Check if correct original url is returned
	s.Equal(resp.OriginalURL, valueFromDB.OriginalURL)

	// Since cache is used, check if value is being saved
	s.Equal(resp.OriginalURL, valueFromCache)
}

func (s *Suite) TestRedirect_NotFound() {
	ctx := context.Background()
	query := queries.RedirectQuery{
		ShortURL: "SOMEURL",
	}

	handler, err := queries.NewRedirectQueryHandler(
		zaptest.NewLogger(s.T()).Sugar(), s.cache, s.pgxPool)
	s.Require().NoError(err)

	resp, handlerErr := handler.Handle(ctx, query)

	valueFromDB, repoErr := s.urlRepo.GetByShortenedURL(ctx, query.ShortURL)

	valueFromCache, cacheErr := s.cache.Get(ctx, query.ShortURL)
	s.Require().NoError(cacheErr)

	// Check if returned error is ErrObjectNotFound
	s.Require().ErrorIs(repoErr, errs.ErrObjectNotFound)
	s.Require().ErrorIs(handlerErr, errs.ErrObjectNotFound)
	s.Empty(resp)
	s.Nil(valueFromDB)
	// Check empty value since request for non-existing values are cached too
	s.Empty(valueFromCache)
}
