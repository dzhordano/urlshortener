package model

import (
	"time"

	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/dzhordano/urlshortener/internal/pkg/random"
)

const ShortURLLength = 8

type ShortenedURL struct {
	OriginalURL  string
	ShortURL     string
	Clicks       int
	CreatedAtUTC time.Time
}

func NewShortenedURL(originalURL string) (*ShortenedURL, error) {
	if originalURL == "" {
		return nil, errs.NewValueIsRequiredError("originalURL")
	}

	return &ShortenedURL{
		OriginalURL:  originalURL,
		ShortURL:     random.NewRandomString(ShortURLLength),
		Clicks:       0,
		CreatedAtUTC: time.Now().UTC(),
	}, nil
}
