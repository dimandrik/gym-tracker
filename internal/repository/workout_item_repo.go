package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkoutItemRepository struct {
	pool *pgxpool.Pool
}

func NewWorkoutItemRepository(pool *pgxpool.Pool) *WorkoutItemRepository {
	return &WorkoutItemRepository{pool}
}

// одна запись на пару "тренировка-машина"; та же гонка при конкурентных запросах, что и в GetOrCreateWorkout
func (r *WorkoutItemRepository) GetOrCreateWorkoutItem(ctx context.Context, workoutID, machineID string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		"SELECT id FROM workout_items WHERE workout_id = $1 AND machine_id = $2",
		workoutID, machineID,
	).Scan(&id)

	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("failed to query workout item: %w", err)
	}

	err = r.pool.QueryRow(ctx,
		"INSERT INTO workout_items (workout_id, machine_id) VALUES ($1, $2) RETURNING id",
		workoutID, machineID,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create workout item: %w", err)
	}
	return id, nil
}
