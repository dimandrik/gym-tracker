package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// требует поднятую тестовую БД с прогнанными миграциями (см. docker-compose.yml);
// при недоступности БД тест пропускается, а не падает, чтобы не ломать `go test ./...` без Postgres
func setupTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = "postgres://gym_user:gym_pass@localhost:5432/gym_tracker"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("skipping: cannot create pool for test database: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skipping: test database not reachable at %s: %v", dsn, err)
	}

	return pool
}

// создаёт пользователя с уникальным email и регистрирует его удаление по завершении теста;
// каскадные FK (миграция 003) подчищают заодно все его тренажёры/тренировки/подходы
func createTestUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()

	ctx := context.Background()
	email := "test-" + uuid.New().String() + "@example.com"

	var userID string
	err := pool.QueryRow(ctx,
		"INSERT INTO users (email, password_hash, first_name, last_name) VALUES ($1, 'x', 'Test', 'User') RETURNING id",
		email,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	})

	return userID
}
