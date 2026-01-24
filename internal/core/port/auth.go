package port

import (
	"context"
	"github.com/yourname/ticketing-system/internal/core/entity"
)

type UserRepositoryPort interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
}

type AuthServicePort interface {
	Register(ctx context.Context, req entity.RegisterRequest) (*entity.User, error)
	Login(ctx context.Context, req entity.LoginRequest) (string, error)
	ValidateToken(ctx context.Context, token string) (*entity.User, error)
}