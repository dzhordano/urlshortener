package model

import (
	"time"

	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/dzhordano/urlshortener/internal/pkg/random"
	"github.com/google/uuid"
)

const (
	ShortURLLength   = 8
	ShortURLValidFor = 14 * 24 * time.Hour // 2 weeks
)

type ShortenedURL struct {
	ID            uuid.UUID
	OriginalURL   string
	ShortURL      string
	Clicks        int
	CreatedAtUTC  time.Time
	ValidUntilUTC time.Time
}

func NewShortenedURL(originalURL string) (*ShortenedURL, error) {
	if originalURL == "" {
		return nil, errs.NewValueIsRequiredError("originalURL")
	}

	n := time.Now()

	return &ShortenedURL{
		ID:            uuid.New(),
		OriginalURL:   originalURL,
		ShortURL:      random.NewRandomString(ShortURLLength),
		Clicks:        0,
		CreatedAtUTC:  n.UTC(),
		ValidUntilUTC: n.Add(ShortURLValidFor).UTC(),
	}, nil
}
