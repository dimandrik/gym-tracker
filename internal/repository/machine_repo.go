package repository

import (
	"context"
	"fmt"
	"gym_tracker/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MachineRepository struct {
	pool *pgxpool.Pool
}

func NewMachineRepository(pool *pgxpool.Pool) *MachineRepository {
	return &MachineRepository{pool}
}

func (r *MachineRepository) CreateMachine(ctx context.Context, userID, name, photoURL string) (string, error) {
	row := r.pool.QueryRow(ctx, "INSERT INTO machines (user_id, name, photo_url) VALUES ($1, $2, $3) RETURNING id", userID, name, photoURL)
	var id string
	err := row.Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create machine: %w", err)
	}
	return id, nil
}

func (r *MachineRepository) GetMachinesByUserID(ctx context.Context, userID string) ([]models.Machine, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, user_id, name, photo_url, created_at FROM machines WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get machines by user ID: %w", err)
	}
	defer rows.Close()

	var machines []models.Machine
	for rows.Next() {
		var m models.Machine
		if err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.PhotoURL, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}
		machines = append(machines, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return machines, nil
}

func (r *MachineRepository) GetMachineByID(ctx context.Context, id string) (*models.Machine, error) {
	row := r.pool.QueryRow(ctx, "SELECT id, user_id, name, photo_url, created_at FROM machines WHERE id = $1", id)

	var m models.Machine
	err := row.Scan(&m.ID, &m.UserID, &m.Name, &m.PhotoURL, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine by ID: %w", err)
	}
	return &m, nil
}
