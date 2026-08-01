package repository

import (
	"context"
	"errors"
	"fmt"
	"gym_tracker/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkoutRepository struct {
	pool *pgxpool.Pool
}

func NewWorkoutRepository(pool *pgxpool.Pool) *WorkoutRepository {
	return &WorkoutRepository{pool}
}

// GetOrCreateWorkout returns the user's workout row for the given date,
// creating one if it doesn't exist yet — this is what lets all sets logged
// on the same day land in a single workout without the frontend having to
// track a "current workout" explicitly.
//
// Not wrapped in a transaction, so two near-simultaneous first-set-of-the-day
// requests could in theory both miss the SELECT and both INSERT a duplicate
// workout row for that date. Not a concern in practice with one client per user.
func (r *WorkoutRepository) GetOrCreateWorkout(ctx context.Context, userID string, date time.Time) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		"SELECT id FROM workouts WHERE user_id = $1 AND workout_date = $2",
		userID, date,
	).Scan(&id)

	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("failed to query workout: %w", err)
	}

	err = r.pool.QueryRow(ctx,
		"INSERT INTO workouts (user_id, workout_date) VALUES ($1, $2) RETURNING id",
		userID, date,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create workout: %w", err)
	}
	return id, nil
}

// GetWorkoutHistory returns one row per workout day with counts of distinct
// machines used and total sets logged — powers the history screen list.
func (r *WorkoutRepository) GetWorkoutHistory(ctx context.Context, userID string) ([]models.WorkoutSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			workouts.workout_date,
			COUNT(DISTINCT workout_items.machine_id) AS machines_count,
			COUNT(sets.id) AS sets_count
		FROM workouts
		JOIN workout_items ON workout_items.workout_id = workouts.id
		JOIN sets ON sets.workout_item_id = workout_items.id
		WHERE workouts.user_id = $1
		GROUP BY workouts.workout_date
		ORDER BY workouts.workout_date DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout history: %w", err)
	}
	defer rows.Close()
	var summaries []models.WorkoutSummary
	for rows.Next() {
		var s models.WorkoutSummary
		if err := rows.Scan(&s.WorkoutDate, &s.MachinesCount, &s.SetsCount); err != nil {
			return nil, fmt.Errorf("failed to scan workout summary: %w", err)
		}
		summaries = append(summaries, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return summaries, nil
}
