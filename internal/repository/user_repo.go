package repository

import (
	"context"
	"errors"
	"fmt"
	"gym_tracker/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (string, error) {
	row := r.pool.QueryRow(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", email, passwordHash)
	var id string
	err := row.Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		// 23505 = unique_violation — relies on a unique constraint on users.email
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrEmailAlreadyExists
		}
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return id, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	row := r.pool.QueryRow(ctx, "SELECT id, email, password_hash FROM users WHERE email = $1", email)
	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &u, nil
}
