package repository

import (
	"context"
	"database/sql"

	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) port.UserRepositoryPort {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	user := &entity.User{}
	query := `SELECT id, username, email, password_hash, role FROM users WHERE email = $1 LIMIT 1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	user := &entity.User{}
	query := `SELECT id, username, email, role FROM users WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}
