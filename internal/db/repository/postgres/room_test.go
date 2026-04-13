package postgres

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"room-booking/internal/domain"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func runMigrations(databaseURL string, migrationsPath string) error {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to resolve migrations path: %w", err)
	}

	migrateURL := strings.Replace(databaseURL, "postgres://", "pgx5://", 1)

	m, err := migrate.New("file://"+absPath, migrateURL)
	if err != nil {
		return fmt.Errorf("failed to init migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(10*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(ctx))
	t.Cleanup(pool.Close)

	migrationsPath := filepath.Join("..", "..", "..", "db", "migrations")
	require.NoError(t, runMigrations(connStr, migrationsPath))

	return pool
}

func TestRoomRepository_CreateAndGet(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewRoomRepository(db)

	ctx := context.Background()
	room, err := domain.NewRoom("Test Room", nil, nil)
	require.NoError(t, err)

	require.NoError(t, repo.Create(ctx, room))

	got, err := repo.GetByID(ctx, room.ID)
	require.NoError(t, err)
	require.Equal(t, room.Name, got.Name)
	require.Equal(t, room.ID, got.ID)
}
