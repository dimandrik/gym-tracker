package repository

import (
	"context"
	"testing"
)

func TestMachineRepository_GetMachineByID_OwnMachine(t *testing.T) {
	pool := setupTestPool(t)
	defer pool.Close()
	repo := NewMachineRepository(pool)
	ctx := context.Background()

	userID := createTestUser(t, pool)
	machineID, err := repo.CreateMachine(ctx, userID, "Leg Press", "/uploads/leg-press.jpg")
	if err != nil {
		t.Fatalf("CreateMachine failed: %v", err)
	}

	machine, err := repo.GetMachineByID(ctx, machineID, userID)
	if err != nil {
		t.Fatalf("expected owner to fetch their own machine, got error: %v", err)
	}
	if machine.ID != machineID {
		t.Errorf("expected machine ID %q, got %q", machineID, machine.ID)
	}
}

func TestMachineRepository_GetMachineByID_OtherUsersMachine(t *testing.T) {
	pool := setupTestPool(t)
	defer pool.Close()
	repo := NewMachineRepository(pool)
	ctx := context.Background()

	owner := createTestUser(t, pool)
	attacker := createTestUser(t, pool)

	machineID, err := repo.CreateMachine(ctx, owner, "Leg Press", "/uploads/leg-press.jpg")
	if err != nil {
		t.Fatalf("CreateMachine failed: %v", err)
	}

	if _, err := repo.GetMachineByID(ctx, machineID, attacker); err == nil {
		t.Fatal("expected error when fetching another user's machine, got nil (IDOR)")
	}
}

func TestMachineRepository_GetMachineByID_NonExistent(t *testing.T) {
	pool := setupTestPool(t)
	defer pool.Close()
	repo := NewMachineRepository(pool)
	ctx := context.Background()

	userID := createTestUser(t, pool)

	if _, err := repo.GetMachineByID(ctx, "00000000-0000-0000-0000-000000000000", userID); err == nil {
		t.Fatal("expected error for non-existent machine, got nil")
	}
}

func TestMachineRepository_DeleteMachine_OtherUsersMachine(t *testing.T) {
	pool := setupTestPool(t)
	defer pool.Close()
	repo := NewMachineRepository(pool)
	ctx := context.Background()

	owner := createTestUser(t, pool)
	attacker := createTestUser(t, pool)

	machineID, err := repo.CreateMachine(ctx, owner, "Leg Press", "/uploads/leg-press.jpg")
	if err != nil {
		t.Fatalf("CreateMachine failed: %v", err)
	}

	if err := repo.DeleteMachine(ctx, machineID, attacker); err == nil {
		t.Fatal("expected error when deleting another user's machine, got nil (IDOR)")
	}

	if _, err := repo.GetMachineByID(ctx, machineID, owner); err != nil {
		t.Fatalf("machine should still exist for the real owner after the rejected delete: %v", err)
	}
}