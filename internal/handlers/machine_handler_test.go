package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gym_tracker/internal/middleware"
	"gym_tracker/internal/repository"
)

func TestMachineHandler_GetMachine_RejectsAnotherUsersMachine(t *testing.T) {
	pool := setupTestPool(t)
	defer pool.Close()

	machineRepo := repository.NewMachineRepository(pool)
	h := NewMachineHandler(machineRepo, "uploads")

	owner := createTestUser(t, pool)
	attacker := createTestUser(t, pool)

	machineID, err := machineRepo.CreateMachine(context.Background(), owner, "Squat Rack", "/uploads/squat.jpg")
	if err != nil {
		t.Fatalf("CreateMachine failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/machines/"+machineID, nil)
	req.SetPathValue("id", machineID)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, attacker))
	rec := httptest.NewRecorder()

	h.GetMachine(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d when fetching another user's machine, got %d (IDOR): %s",
			http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestMachineHandler_GetMachine_OwnMachine(t *testing.T) {
	pool := setupTestPool(t)
	defer pool.Close()

	machineRepo := repository.NewMachineRepository(pool)
	h := NewMachineHandler(machineRepo, "uploads")

	userID := createTestUser(t, pool)
	machineID, err := machineRepo.CreateMachine(context.Background(), userID, "Squat Rack", "/uploads/squat.jpg")
	if err != nil {
		t.Fatalf("CreateMachine failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/machines/"+machineID, nil)
	req.SetPathValue("id", machineID)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	rec := httptest.NewRecorder()

	h.GetMachine(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d when fetching own machine, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
}