package urlcache

import (
	"context"
	"errors"
	"time"

	"github.com/dzhordano/urlshortener/internal/core/ports"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewRedisCache(rdb *redis.Client, ttl time.Duration) (ports.URLCache, error) {
	if rdb == nil {
		return nil, errs.NewValueIsRequiredError("rdb")
	}

	return &Cache{rdb: rdb, ttl: ttl}, nil
}

func (c *Cache) Set(ctx context.Context, key string, value string) error {
	sc := c.rdb.Set(ctx, key, value, c.ttl)
	if sc.Err() != nil {
		return sc.Err()
	}

	return nil
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	sc := c.rdb.Get(ctx, key)
	if sc.Err() != nil {
		if errors.Is(sc.Err(), redis.Nil) {
			return "", errs.NewObjectNotFoundError("key", key)
		}
		return "", sc.Err()
	}

	return sc.Val(), nil
}
