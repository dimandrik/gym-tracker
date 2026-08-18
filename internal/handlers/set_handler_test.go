package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"gym_tracker/internal/middleware"
	"gym_tracker/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// требует поднятую тестовую БД с прогнанными миграциями; при недоступности БД — скип
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

func TestSetHandler_AddSet_RejectsAnotherUsersMachine(t *testing.T) {
	pool := setupTestPool(t)
	defer pool.Close()

	machineRepo := repository.NewMachineRepository(pool)
	workoutRepo := repository.NewWorkoutRepository(pool)
	workoutItemRepo := repository.NewWorkoutItemRepository(pool)
	setRepo := repository.NewSetRepository(pool)
	h := NewSetHandler(workoutRepo, workoutItemRepo, setRepo, machineRepo)

	owner := createTestUser(t, pool)
	attacker := createTestUser(t, pool)

	machineID, err := machineRepo.CreateMachine(context.Background(), owner, "Squat Rack", "/uploads/squat.jpg")
	if err != nil {
		t.Fatalf("CreateMachine failed: %v", err)
	}

	body, _ := json.Marshal(AddSetRequest{MachineID: machineID, WeightKg: 100, Reps: 5})
	req := httptest.NewRequest(http.MethodPost, "/api/sets", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, attacker))
	rec := httptest.NewRecorder()

	h.AddSet(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d when adding a set to another user's machine, got %d (IDOR): %s",
			http.StatusNotFound, rec.Code, rec.Body.String())
	}

	sets, err := setRepo.GetSetsByMachineID(context.Background(), machineID, owner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sets) != 0 {
		t.Fatalf("expected no set to be created against another user's machine, got %d", len(sets))
	}
}

func TestSetHandler_AddSet_OwnMachine(t *testing.T) {
	pool := setupTestPool(t)
	defer pool.Close()

	machineRepo := repository.NewMachineRepository(pool)
	workoutRepo := repository.NewWorkoutRepository(pool)
	workoutItemRepo := repository.NewWorkoutItemRepository(pool)
	setRepo := repository.NewSetRepository(pool)
	h := NewSetHandler(workoutRepo, workoutItemRepo, setRepo, machineRepo)

	userID := createTestUser(t, pool)
	machineID, err := machineRepo.CreateMachine(context.Background(), userID, "Squat Rack", "/uploads/squat.jpg")
	if err != nil {
		t.Fatalf("CreateMachine failed: %v", err)
	}

	body, _ := json.Marshal(AddSetRequest{MachineID: machineID, WeightKg: 100, Reps: 5})
	req := httptest.NewRequest(http.MethodPost, "/api/sets", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	rec := httptest.NewRecorder()

	h.AddSet(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected %d when adding a set to own machine, got %d: %s",
			http.StatusCreated, rec.Code, rec.Body.String())
	}
}