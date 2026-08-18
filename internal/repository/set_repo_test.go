package repository

import (
	"context"
	"testing"
	"time"
)

func TestSetRepository_GetSetsByMachineID_OwnMachine(t *testing.T) {
	pool := setupTestPool(t)
	defer pool.Close()
	ctx := context.Background()

	machineRepo := NewMachineRepository(pool)
	workoutRepo := NewWorkoutRepository(pool)
	workoutItemRepo := NewWorkoutItemRepository(pool)
	setRepo := NewSetRepository(pool)

	userID := createTestUser(t, pool)
	machineID, err := machineRepo.CreateMachine(ctx, userID, "Bench Press", "/uploads/bench.jpg")
	if err != nil {
		t.Fatalf("CreateMachine failed: %v", err)
	}
	workoutID, err := workoutRepo.GetOrCreateWorkout(ctx, userID, time.Now().Truncate(24*time.Hour))
	if err != nil {
		t.Fatalf("GetOrCreateWorkout failed: %v", err)
	}
	itemID, err := workoutItemRepo.GetOrCreateWorkoutItem(ctx, workoutID, machineID)
	if err != nil {
		t.Fatalf("GetOrCreateWorkoutItem failed: %v", err)
	}
	if _, err := setRepo.CreateSet(ctx, itemID, 1, 60, 10); err != nil {
		t.Fatalf("CreateSet failed: %v", err)
	}

	sets, err := setRepo.GetSetsByMachineID(ctx, machineID, userID)
	if err != nil {
		t.Fatalf("expected owner to fetch sets for their own machine, got error: %v", err)
	}
	if len(sets) != 1 {
		t.Fatalf("expected 1 set, got %d", len(sets))
	}
}

func TestSetRepository_GetSetsByMachineID_OtherUsersMachine(t *testing.T) {
	pool := setupTestPool(t)
	defer pool.Close()
	ctx := context.Background()

	machineRepo := NewMachineRepository(pool)
	workoutRepo := NewWorkoutRepository(pool)
	workoutItemRepo := NewWorkoutItemRepository(pool)
	setRepo := NewSetRepository(pool)

	owner := createTestUser(t, pool)
	attacker := createTestUser(t, pool)

	machineID, err := machineRepo.CreateMachine(ctx, owner, "Bench Press", "/uploads/bench.jpg")
	if err != nil {
		t.Fatalf("CreateMachine failed: %v", err)
	}
	workoutID, err := workoutRepo.GetOrCreateWorkout(ctx, owner, time.Now().Truncate(24*time.Hour))
	if err != nil {
		t.Fatalf("GetOrCreateWorkout failed: %v", err)
	}
	itemID, err := workoutItemRepo.GetOrCreateWorkoutItem(ctx, workoutID, machineID)
	if err != nil {
		t.Fatalf("GetOrCreateWorkoutItem failed: %v", err)
	}
	if _, err := setRepo.CreateSet(ctx, itemID, 1, 60, 10); err != nil {
		t.Fatalf("CreateSet failed: %v", err)
	}

	// без ошибки, но пустой результат — запрос отфильтровал по чужому machines.user_id
	sets, err := setRepo.GetSetsByMachineID(ctx, machineID, attacker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sets) != 0 {
		t.Fatalf("expected no sets to leak to another user, got %d (IDOR)", len(sets))
	}
}