package cmd

import (
	"github.com/dzhordano/urlshortener/internal/adapters/outbound/pg/urlrepo"
	"github.com/dzhordano/urlshortener/internal/adapters/outbound/redis/urlcache"
	"github.com/dzhordano/urlshortener/internal/core/application/usecases/commands"
	"github.com/dzhordano/urlshortener/internal/core/application/usecases/queries"
	"github.com/dzhordano/urlshortener/internal/core/ports"
	"github.com/dzhordano/urlshortener/internal/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type CompositionRoot struct {
	log logger.Logger
	cfg Config
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewCompositionRoot(
	log logger.Logger,
	cfg Config,
	db *pgxpool.Pool,
	rdb *redis.Client,
) *CompositionRoot {
	return &CompositionRoot{
		log: log,
		cfg: cfg,
		db:  db,
		rdb: rdb,
	}
}

func (cr *CompositionRoot) NewURLRepository() ports.URLRepository {
	urlRepo, err := urlrepo.NewRepository(cr.db)
	if err != nil {
		cr.log.Panicw("error creating url repo", "error", err)
	}
	return urlRepo
}

func (cr *CompositionRoot) NewURLCache() ports.URLCache {
	cache, err := urlcache.NewRedisCache(
		cr.rdb,
		cr.cfg.RDB.TTL,
	)
	if err != nil {
		cr.log.Panicw("error creating redis cache", "error", err)
	}
	return cache
}

func (cr *CompositionRoot) NewShortenURLCommandHandler() commands.ShortenURLCommandHandler {
	handler, err := commands.NewShortenURLCommandHandler(
		cr.log,
		cr.NewURLCache(),
		cr.NewURLRepository(),
	)
	if err != nil {
		cr.log.Panicw("error creating shorten url command handler", "error", err)
	}

	return handler
}

func (cr *CompositionRoot) NewRedirectQueryHandler() queries.RedirectQueryHandler {
	handler, err := queries.NewRedirectQueryHandler(cr.log, cr.NewURLCache(), cr.db)
	if err != nil {
		cr.log.Panicw("error creating redirect query handler", "error", err)
	}

	return handler
}

func (cr *CompositionRoot) NewGetURLInfoQueryHandler() queries.GetURLInfoQueryHandler {
	handler, err := queries.NewGetURLInfoQueryHandler(cr.log, cr.db)
	if err != nil {
		cr.log.Panicw("error creating get url info query handler", "error", err)
	}
	return handler
}
