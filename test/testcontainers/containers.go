package testcontainers

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	// pgx stdlib driver for goose migrations
	_ "github.com/jackc/pgx/v5/stdlib"
	"database/sql"
)

// TestContainers holds running test containers and their clients.
type TestContainers struct {
	PostgresContainer testcontainers.Container
	RedisContainer    testcontainers.Container
	Database          *pgxpool.Pool
	RedisClient       *redis.Client
}

// SetupTestContainers spins up Postgres + Redis containers, runs migrations,
// and returns clients ready for testing.
func SetupTestContainers(t *testing.T) *TestContainers {
	t.Helper()
	ctx := context.Background()

	// Start Postgres
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16.2-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       "booker_test",
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")
	pgURI := fmt.Sprintf("postgresql://postgres:postgres@%s:%s/booker_test?sslmode=disable", pgHost, pgPort.Port())

	// Run migrations
	runMigrations(t, pgURI)

	// Create pgx pool
	pool, err := pgxpool.New(ctx, pgURI)
	if err != nil {
		t.Fatalf("failed to create pgx pool: %v", err)
	}

	// Start Redis
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7.2-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	redisHost, _ := redisContainer.Host(ctx)
	redisPort, _ := redisContainer.MappedPort(ctx, "6379")
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort.Port()),
	})

	tc := &TestContainers{
		PostgresContainer: pgContainer,
		RedisContainer:    redisContainer,
		Database:          pool,
		RedisClient:       redisClient,
	}

	t.Cleanup(func() {
		pool.Close()
		redisClient.Close()
		_ = pgContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
	})

	return tc
}

func runMigrations(t *testing.T, dbURI string) {
	t.Helper()

	db, err := sql.Open("pgx", dbURI)
	if err != nil {
		t.Fatalf("failed to open db for migrations: %v", err)
	}
	defer db.Close()

	migrationsDir := filepath.Join(projectRoot(), "migrations")
	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
}

// projectRoot returns the absolute path to the project root.
func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// test/testcontainers/containers.go → ../../
	return filepath.Join(filepath.Dir(filename), "..", "..")
}
