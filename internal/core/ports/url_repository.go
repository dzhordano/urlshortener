package ports

import (
	"context"

	"github.com/dzhordano/urlshortener/internal/core/domain/model"
)

type URLRepository interface {
	Save(ctx context.Context, url *model.ShortenedURL) error
	GetByShortenedURL(ctx context.Context, shortenedURL string) (*model.ShortenedURL, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (*model.ShortenedURL, error)
}
