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

func (r *SetRepository) GetSetsByMachineID(ctx context.Context, machineID, userID string) ([]models.SetHistoryEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT sets.id, machines.name, workouts.workout_date, sets.set_number, sets.weight_kg, sets.reps
		FROM sets
		JOIN workout_items ON sets.workout_item_id = workout_items.id
		JOIN machines ON workout_items.machine_id = machines.id
		JOIN workouts ON workout_items.workout_id = workouts.id
		WHERE machines.id = $1 AND machines.user_id = $2
		ORDER BY workouts.workout_date DESC, sets.set_number
	`, machineID, userID)
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

// после удаления сдвигает set_number последующих подходов на 1 назад, чтобы нумерация оставалась без разрывов
func (r *SetRepository) DeleteSet(ctx context.Context, setID, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

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

// номер = количество подходов + 1, без отдельного счётчика; при параллельных запросах
// возможна гонка, но это не проблема при одном устройстве на пользователя
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
