package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dzhordano/urlshortener/cmd"
	http_inbound "github.com/dzhordano/urlshortener/internal/adapters/inbound/httpinbound"
	"github.com/dzhordano/urlshortener/internal/core/application/usecases/commands"
	"github.com/dzhordano/urlshortener/internal/core/application/usecases/queries"
	"github.com/dzhordano/urlshortener/internal/gen/servers"
	"github.com/dzhordano/urlshortener/internal/pkg/logger"
	"github.com/dzhordano/urlshortener/migrations"
	echoPrometheus "github.com/globocom/echo-prometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/oklog/run"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
)

//nolint:gocognit,funlen // TODO For now suppress, but actually must be optimized.
func main() {
	// Load .env file.
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	// Load config.
	cfg := mustLoadFromEnv()

	// Handling log level using env, not cool.
	ifDev := cfg.Environment == "dev"
	var logLevel string
	switch cfg.Environment {
	case "dev":
		logLevel = "debug"
	case "prod":
		logLevel = "error"
	default:
		logLevel = "info"
	}

	l, err := logger.NewSlogLogger(ifDev, logLevel)
	if err != nil {
		log.Fatalf("failed to init slog logger: %v", err)
	}

	cr := cmd.NewCompositionRoot(l, cfg)

	ctx, cancel := context.WithCancel(context.Background())

	pool, err := newPgxPool(ctx, cfg.DB.DSN())
	if err != nil {
		log.Fatalf("error creating pgxpool: %v", err)
	}
	cr.RegisterCloseFn(func(ctx context.Context) error {
		pool.Close()
		return nil
	})

	// Open sql.DB for migration purposes.
	db, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		log.Fatalf("error opening sql.DB: %v", err)
	}

	// Set migrations FS for goose.
	goose.SetBaseFS(migrations.FS)

	// Apply migrations
	gooseMigrate(db, ".")
	if err = db.Close(); err != nil {
		l.Warn("error closing sql.DB", "error", err)
	}

	rdb := newRedisClient(cfg.RDB.Addr(), cfg.RDB.Password)
	cr.RegisterCloseFn(func(ctx context.Context) error {
		return rdb.Close()
	})

	urlCache := cr.NewURLCache(rdb)
	urlRepo := cr.NewURLRepository(pool)

	// Init otel tracer.
	tp, err := initTracer(ctx, cfg.ServiceName, cfg.JaegerURL, cfg.Environment)
	if err != nil {
		log.Fatalf("error initializing tracer: %v", err)
	}
	cr.RegisterCloseFn(func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	})

	e := newEchoWebServer(
		cfg.ServiceName,
		cr.NewShortenURLCommandHandler(urlCache, urlRepo),
		cr.NewRedirectQueryHandler(urlCache, pool),
		cr.NewGetURLInfoQueryHandler(pool),
	)

	cs, err := cr.NewCronScheduler()
	if err != nil {
		log.Fatalf("failed to create cron scheduler: %v", err)
	}
	cr.RegisterCloseFn(func(ctx context.Context) error {
		return cs.Stop(ctx)
	})

	cleanExpURLsTask, err := cr.NewCleanExpiredURLsCronTask(pool)
	if err != nil {
		log.Fatalf("failed to create cron task for cleaning expired urls: %v", err)
	}

	// Schedule so cleaning happens every 5 minutes.
	// TODO Hardcoded, could move to config.
	if err := cs.ScheduleInterval(ctx, cleanExpURLsTask, 5*time.Minute); err != nil {
		log.Fatalf("failed to schedule cron task for cleaning expired urls: %v", err)
	}

	// Using run.Group handle startup and graceful shutdown. pretti usful.
	var g run.Group

	// Run app.
	g.Add(func() error {
		l.Info("starting echo server", "address", cfg.HTTP.Addr())

		if err := cs.Start(ctx); err != nil {
			return err
		}

		return e.Start(cfg.HTTP.Addr())
	}, func(error) {
		l.Info("shutting down http server")
		//nolint:mnd // TODO Magic number ahead.
		ctxShutdown, wsCancel := context.WithTimeout(ctx, 10*time.Second)
		defer wsCancel()
		if sdErr := e.Shutdown(ctxShutdown); sdErr != nil {
			l.Warn("failed to shutdown echo server", "error", sdErr)
			if cErr := e.Close(); cErr != nil {
				l.Warn("failed to close echo server", "error", cErr)
			}
		}
	})

	// Shutdown channel. Could've also used context with signals though.
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Listen for signals.
	g.Add(func() error {
		select {
		case sig := <-s:
			l.Info("received signal", "signal", sig)
			return fmt.Errorf("received signal %s", sig)
		case <-ctx.Done():
			return ctx.Err()
		}
	}, func(error) {
		cancel()
		signal.Stop(s)
		close(s)
	})

	// Free resources.
	g.Add(func() error {
		<-ctx.Done()
		return nil
	}, func(error) {
		if cErr := cr.CloseWithTimeout(30 * time.Second); cErr != nil {
			l.Warn("failed to close resources", "error", err)
		}
	})

	// Run everything.
	if rgErr := g.Run(); rgErr != nil {
		l.Warn("run group stopped", "error", rgErr)
	}
}

// Kind of primitive config parse.
func mustLoadFromEnv() cmd.Config {
	rdbttl, err := time.ParseDuration(os.Getenv("REDIS_TTL"))
	if err != nil {
		log.Fatalf("error parsing redis ttl: %v", err)
	}

	return cmd.Config{
		Environment: os.Getenv("ENVIRONMENT"),
		ServiceName: os.Getenv("SERVICE_NAME"),
		HTTP: cmd.HTTPConfig{
			Host: os.Getenv("HTTP_HOST"),
			Port: os.Getenv("HTTP_PORT"),
		},
		DB: cmd.DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
		},
		RDB: cmd.RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
			TTL:      rdbttl,
		},
		JaegerURL: os.Getenv("JAEGER_URL"),
	}
}

func newPgxPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	//nolint:mnd // TODO Magic number ahead.
	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = pool.Ping(timeout)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func newRedisClient(addr string, pass string) *redis.Client {
	//nolint:exhaustruct // Not necessary?
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
	})

	//nolint:mnd // TODO Magic number ahead.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// hope it works with ctx timeout...?
	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Printf("error connecting to redis: %v", err)
	}

	return rdb
}

// Applies every migrations file from dir. Could use [embed] fs, however must use [goose.SetBaseFS].
func gooseMigrate(db *sql.DB, dir string) {
	// Remove* goose logger
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("error setting goose dialect: %v", err)
	}

	if err := goose.Up(db, dir); err != nil {
		log.Fatalf("error migrating db: %v", err)
	}
}

//nolint:mnd // FIXME Magic numbers.
func initTracer(
	ctx context.Context,
	serviceName, jaegerURL, environment string,
) (*sdktrace.TracerProvider, error) {
	exp, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(jaegerURL),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithTimeout(5*time.Second),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		otlptracehttp.WithRetry(
			otlptracehttp.RetryConfig{
				Enabled:         true,
				InitialInterval: 1 * time.Second,
				MaxInterval:     5 * time.Second,
				MaxElapsedTime:  30 * time.Second,
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Init resource with service metadata
	attrResource, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version.Version),
			attribute.String("environment", environment),
		),
		resource.WithOS(),
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(attrResource),
		// Sample only parent based spans
		// Save all spans. TODO could make this more flexible but eh...
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1.0))),
	)

	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))

	return tracerProvider, nil
}

func newEchoWebServer(
	tracerServerName string,
	shortenCHandler commands.ShortenURLCommandHandler,
	redirectQHandler queries.RedirectQueryHandler,
	getUrlInfoQHandler queries.GetURLInfoQueryHandler,
) *echo.Echo {
	e := echo.New()

	e.Use(
		// Setup otel middleware for tracing.
		otelecho.Middleware(
			tracerServerName,
			// Skip tracing for /metrics endpoint
			otelecho.WithSkipper(
				func(c echo.Context) bool {
					return c.Path() == "/metrics"
				}),
		),
		// Use echo rate limiter. Pretty lame, but for <= 100 rps is supposedly okay.
		middleware.RateLimiter(
			//nolint:mnd // Using 100 rps as default.
			middleware.NewRateLimiterMemoryStore(100),
		),
		// Recover panics.
		middleware.Recover(),
		// Use default CORS settings.
		middleware.CORS(),
	)

	handlers, err := http_inbound.NewServer(
		shortenCHandler,
		redirectQHandler,
		getUrlInfoQHandler,
	)
	if err != nil {
		log.Fatalf("error creating server: %v", err)
	}

	registerMetrics(e)
	registerSwagOpenAPI(e)
	registerSwagUI(e)
	servers.RegisterHandlers(e, handlers)

	return e
}

func registerSwagOpenAPI(e *echo.Echo) {
	e.GET("/openapi.json", func(c echo.Context) error {
		s, err := servers.GetSwagger()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		b, err := s.MarshalJSON()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		return c.JSONBlob(http.StatusOK, b)
	})
}

// Copied from swagger site.
func registerSwagUI(e *echo.Echo) {
	e.GET("/docs", func(c echo.Context) error {
		html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
		  <meta charset="utf-8" />
		  <meta name="viewport" content="width=device-width, initial-scale=1" />
		  <meta name="description" content="SwaggerUI" />
		  <title>SwaggerUI</title>
		  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
		</head>
		<body>
		<div id="swagger-ui"></div>
		<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
		<script>
		  window.onload = () => {
		    window.ui = SwaggerUIBundle({
		      url: '/openapi.json',
		      dom_id: '#swagger-ui',
		    });
		  };
		</script>
		</body>
		</html>`
		return c.HTML(http.StatusOK, html)
	})
}

func registerMetrics(e *echo.Echo) {
	// Using default config
	e.Use(echoPrometheus.MetricsMiddleware())
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}
