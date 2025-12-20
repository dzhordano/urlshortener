package commands

import (
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
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
