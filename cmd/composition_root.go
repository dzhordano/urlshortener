package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/dzhordano/urlshortener/internal/adapters/outbound/pg/urlrepo"
	"github.com/dzhordano/urlshortener/internal/adapters/outbound/redis/urlcache"
	"github.com/dzhordano/urlshortener/internal/core/application/usecases/commands"
	"github.com/dzhordano/urlshortener/internal/core/application/usecases/queries"
	"github.com/dzhordano/urlshortener/internal/core/ports"
	"github.com/dzhordano/urlshortener/internal/pkg/logger"
	"github.com/dzhordano/urlshortener/internal/pkg/scheduler"
	"github.com/dzhordano/urlshortener/internal/pkg/scheduler/cron"
	"github.com/dzhordano/urlshortener/internal/pkg/scheduler/tasks"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type CloseFn func(context.Context) error

type CompositionRoot struct {
	log logger.Logger
	cfg Config

	closeFns []CloseFn
}

func NewCompositionRoot(
	log logger.Logger,
	cfg Config,
) *CompositionRoot {
	return &CompositionRoot{
		log: log,
		cfg: cfg,
	}
}

// RegisterCloseFn register CloseFn function to be executed upon app's shutdown.
//
// CloseFn's will be called in LIFO manner.
func (cr *CompositionRoot) RegisterCloseFn(closer CloseFn) {
	cr.closeFns = append(cr.closeFns, closer)
}

// Close executes CloseFn's functions.
func (cr *CompositionRoot) Close(ctx context.Context) error {
	var errs []error

	// closing from last to first
	for i := len(cr.closeFns) - 1; i >= 0; i-- {
		closeFn := cr.closeFns[i]
		if err := closeFn(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%T: %w", closeFn, err))
			cr.log.Error("failed to close resource", "error", err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	cr.log.Info("application stopped gracefully")
	return nil
}

func (cr *CompositionRoot) CloseWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return cr.Close(ctx)
}

func (cr *CompositionRoot) NewURLRepository(db *pgxpool.Pool) ports.URLRepository {
	urlRepo, err := urlrepo.NewRepository(db)
	if err != nil {
		cr.log.Error("error creating url repo", "error", err)
	}
	return urlRepo
}

func (cr *CompositionRoot) NewURLCache(rdb *redis.Client) ports.URLCache {
	cache, err := urlcache.NewRedisCache(
		rdb,
		cr.cfg.RDB.TTL,
	)
	if err != nil {
		cr.log.Error("error creating redis cache", "error", err)
	}
	return cache
}

func (cr *CompositionRoot) NewShortenURLCommandHandler(
	urlCache ports.URLCache,
	urlRepo ports.URLRepository,
) commands.ShortenURLCommandHandler {
	handler, err := commands.NewShortenURLCommandHandler(
		cr.log,
		urlCache,
		urlRepo,
	)
	if err != nil {
		cr.log.Error("error creating shorten url command handler", "error", err)
	}

	return handler
}

func (cr *CompositionRoot) NewRedirectQueryHandler(
	urlCache ports.URLCache,
	db *pgxpool.Pool,
) queries.RedirectQueryHandler {
	handler, err := queries.NewRedirectQueryHandler(cr.log, urlCache, db)
	if err != nil {
		cr.log.Error("error creating redirect query handler", "error", err)
	}

	return handler
}

func (cr *CompositionRoot) NewGetURLInfoQueryHandler(
	db *pgxpool.Pool,
) queries.GetURLInfoQueryHandler {
	handler, err := queries.NewGetURLInfoQueryHandler(cr.log, db)
	if err != nil {
		cr.log.Error("error creating get url info query handler", "error", err)
	}
	return handler
}

func (cr *CompositionRoot) NewCronScheduler() (scheduler.Scheduler, error) {
	cs := cron.NewCronScheduler(cr.log)
	return cs, nil
}

func (cr *CompositionRoot) NewCleanExpiredURLsCronTask(
	db *pgxpool.Pool,
) (scheduler.Task, error) {
	cj := tasks.NewCleanupExpiredURLsTask(db)
	return cj, nil
}
