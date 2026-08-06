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

type SetRepository struct {
	pool *pgxpool.Pool
}

func NewSetRepository(pool *pgxpool.Pool) *SetRepository {
	return &SetRepository{pool}
}

func (r *SetRepository) CreateSet(ctx context.Context, workoutItemID string, setNumber int, weightKg float64, reps int) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		"INSERT INTO sets (workout_item_id, set_number, weight_kg, reps) VALUES ($1, $2, $3, $4) RETURNING id",
		workoutItemID, setNumber, weightKg, reps,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create set: %w", err)
	}
	return id, nil
}

// GetSetsByMachineID walks sets -> workout_items -> machines/workouts to
// flatten everything into one row per set, newest workout first. Used for
// the per-machine history screen and for computing the personal record.
func (r *SetRepository) GetSetsByMachineID(ctx context.Context, machineID string) ([]models.SetHistoryEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT sets.id, machines.name, workouts.workout_date, sets.set_number, sets.weight_kg, sets.reps
		FROM sets
		JOIN workout_items ON sets.workout_item_id = workout_items.id
		JOIN machines ON workout_items.machine_id = machines.id
		JOIN workouts ON workout_items.workout_id = workouts.id
		WHERE machines.id = $1
		ORDER BY workouts.workout_date DESC, sets.set_number
	`, machineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sets by machine ID: %w", err)
	}
	defer rows.Close()

	var sets []models.SetHistoryEntry
	for rows.Next() {
		var s models.SetHistoryEntry
		if err := rows.Scan(&s.ID, &s.MachineName, &s.WorkoutDate, &s.SetNumber, &s.WeightKg, &s.Reps); err != nil {
			return nil, fmt.Errorf("failed to scan set: %w", err)
		}
		sets = append(sets, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return sets, nil
}

// DeleteSet removes a set and shifts the set_number of every later set in
// the same workout_item down by one, so numbering stays contiguous from 1
// instead of leaving a gap where the deleted set used to be.
func (r *SetRepository) DeleteSet(ctx context.Context, setID, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var workoutItemID string
	var setNumber int
	err = tx.QueryRow(ctx, `
		DELETE FROM sets
		WHERE id = $1
		AND workout_item_id IN (
			SELECT workout_items.id FROM workout_items
			JOIN workouts ON workout_items.workout_id = workouts.id
			WHERE workouts.user_id = $2
		)
		RETURNING workout_item_id, set_number
	`, setID, userID).Scan(&workoutItemID, &setNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("no set found with id %s for user %s", setID, userID)
		}
		return fmt.Errorf("failed to delete set: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE sets SET set_number = set_number - 1
		WHERE workout_item_id = $1 AND set_number > $2
	`, workoutItemID, setNumber)
	if err != nil {
		return fmt.Errorf("failed to renumber sets: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *SetRepository) UpdateSet(ctx context.Context, setID, userID string, weightKg float64, reps int) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE sets
		SET weight_kg = $1, reps = $2
		WHERE id = $3
		AND workout_item_id IN (
			SELECT workout_items.id FROM workout_items
			JOIN workouts ON workout_items.workout_id = workouts.id
			WHERE workouts.user_id = $4
		)`, weightKg, reps, setID, userID)
	if err != nil {
		return fmt.Errorf("failed to update set: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no set found with id %s for user %s", setID, userID)
	}
	return nil
}

// GetNextSetNumber just counts existing sets + 1 rather than tracking a
// counter — simple, but two concurrent AddSet calls for the same machine
// could both compute the same number. Not an issue with a single user
// logging sets from one device.
func (r *SetRepository) GetNextSetNumber(ctx context.Context, workoutItemID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM sets WHERE workout_item_id = $1",
		workoutItemID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sets: %w", err)
	}
	return count + 1, nil
}

func (r *SetRepository) GetSetByID(ctx context.Context, setID, userID string) (*models.SetDetail, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT sets.id, machines.id, machines.name, workouts.workout_date, sets.set_number, sets.weight_kg, sets.reps
		FROM sets
		JOIN workout_items ON sets.workout_item_id = workout_items.id
		JOIN machines ON workout_items.machine_id = machines.id
		JOIN workouts ON workout_items.workout_id = workouts.id
		WHERE sets.id = $1 AND workouts.user_id = $2
	`, setID, userID)

	var s models.SetDetail
	err := row.Scan(&s.ID, &s.MachineID, &s.MachineName, &s.WorkoutDate, &s.SetNumber, &s.WeightKg, &s.Reps)
	if err != nil {
		return nil, fmt.Errorf("failed to get set by id: %w", err)
	}
	return &s, nil

}

func (r *SetRepository) GetSetsByDate(ctx context.Context, userID string, date time.Time) ([]models.DaySetEntry, error) {
	rows, err := r.pool.Query(ctx, `
	SELECT machines.id, machines.name, machines.photo_url, sets.set_number, sets.weight_kg, sets.reps
	FROM sets
	JOIN workout_items ON sets.workout_item_id = workout_items.id
	JOIN machines ON workout_items.machine_id = machines.id
	JOIN workouts ON workout_items.workout_id = workouts.id
	WHERE workouts.user_id = $1 AND workouts.workout_date = $2
	ORDER BY machines.name, sets.set_number
`, userID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get sets by date: %w", err)
	}
	defer rows.Close()

	var entries []models.DaySetEntry
	for rows.Next() {
		var e models.DaySetEntry
		if err := rows.Scan(&e.MachineID, &e.MachineName, &e.MachinePhotoURL, &e.SetNumber, &e.WeightKg, &e.Reps); err != nil {
			return nil, fmt.Errorf("failed to scan day set entry: %w", err)
		}
		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return entries, nil
}
