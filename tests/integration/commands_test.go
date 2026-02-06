package integration_test

import (
	"context"

	"github.com/dzhordano/urlshortener/internal/core/application/usecases/commands"
)

func (s *Suite) TestShortenURLCommandHandler_Success() {
	ctx := context.Background()
	req := commands.ShortenURLCommand{
		OriginalURL: "http://example.com",
	}

	handler, err := commands.NewShortenURLCommandHandler(s.l, s.cache, s.urlRepo)
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
	s.NotEmpty(valueFromDB.ValidUntilUTC)
}
