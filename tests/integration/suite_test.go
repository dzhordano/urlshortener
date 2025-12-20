package integration_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/dzhordano/urlshortener/internal/adapters/outbound/pg/urlrepo"
	"github.com/dzhordano/urlshortener/internal/adapters/outbound/redis/urlcache"
	"github.com/dzhordano/urlshortener/internal/core/ports"
	"github.com/dzhordano/urlshortener/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

//nolint:embeddedstructfieldcheck // suite.Suite needs to be embedded.
type Suite struct {
	suite.Suite
	pgContainer    *postgres.PostgresContainer
	redisContainer *tcredis.RedisContainer

	pgxPool *pgxpool.Pool
	redisDB *redis.Client

	urlRepo ports.URLRepository
	cache   ports.URLCache
}

func (s *Suite) SetupSuite() {
	start := time.Now()
	defer func() {
		s.T().Logf("SetupSuite completed in %v", time.Since(start))
	}()

	ctx := context.Background()

	// Init postgres container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:18-alpine3.22",
		postgres.WithDatabase("db1"),
		postgres.WithUsername("user1"),
		postgres.WithPassword("password1"),
		postgres.BasicWaitStrategies(),
	)
	s.Require().NoError(err)

	// Init redis container
	redisContainer, err := tcredis.Run(ctx,
		"redis:8.4-alpine3.22",
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
	)
	s.Require().NoError(err)

	dsn, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	// Apply schema migrations and setup sql.DB for db state altering
	db, err := sql.Open("postgres", dsn)
	s.Require().NoError(err)

	goose.SetBaseFS(migrations.FS)
	err = goose.SetDialect("postgres")
	s.Require().NoError(err)

	// Perform migrations
	err = goose.Up(db, ".")
	s.Require().NoError(err)

	err = db.Close()
	s.Require().NoError(err)

	// Init pgxpool
	pool, err := pgxpool.New(ctx, dsn)
	s.Require().NoError(err)

	// Retrieve redis host and port
	rHost, err := redisContainer.Host(ctx)
	s.Require().NoError(err)
	rPort, err := redisContainer.MappedPort(ctx, "6379")
	s.Require().NoError(err)

	// Init redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     rHost + ":" + rPort.Port(),
		Password: "",
		DB:       0,
	})

	timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ping redis db
	err = rdb.Ping(timeout).Err()
	s.Require().NoError(err)

	// Init repositories
	urlRepo, err := urlrepo.NewRepository(pool)
	s.Require().NoError(err)

	c, err := urlcache.NewRedisCache(rdb, 5*time.Second)
	s.Require().NoError(err)

	// Set suite values
	s.pgContainer = postgresContainer
	s.redisContainer = redisContainer
	s.pgxPool = pool
	s.redisDB = rdb
	s.urlRepo = urlRepo
	s.cache = c
}

func (s *Suite) TearDownSuite() {
	var tdErrs []error
	s.pgxPool.Close()

	err := testcontainers.TerminateContainer(s.pgContainer)
	if err != nil {
		tdErrs = append(tdErrs, err)
	}

	err = testcontainers.TerminateContainer(s.redisContainer)
	if err != nil {
		tdErrs = append(tdErrs, err)
	}

	if len(tdErrs) > 0 {
		s.T().Errorf("errors occurred during suite teardown: %v", tdErrs)
	}
}

func (s *Suite) SetupTest() {
	// Empty...
}

func (s *Suite) TearDownTest() {
	// Truncate all tables
	_, err := s.pgxPool.Exec(context.Background(), "TRUNCATE TABLE urls")
	s.NoError(err)

	// Clear redis cache
	s.NoError(s.redisDB.FlushAll(context.Background()).Err())

	// ... these are driver-dependent, thus could move them into separate functions
}

func TestRunSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	s := new(Suite)
	suite.Run(t, s)
}
