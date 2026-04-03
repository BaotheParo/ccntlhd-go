package port

import (
	"context"
	"github.com/yourname/ticketing-system/internal/core/entity"
)

type UserRepositoryPort interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	ListUsers(ctx context.Context, page, limit int) ([]entity.User, int64, error)
	UpdateUserStatus(ctx context.Context, id string, isActive bool) error
}

type AuthServicePort interface {
	Register(ctx context.Context, req entity.RegisterRequest) (*entity.User, error)
	Login(ctx context.Context, req entity.LoginRequest) (string, error)
	ValidateToken(ctx context.Context, token string) (*entity.User, error)
	ListUsers(ctx context.Context, page, limit int) ([]entity.User, int64, error)
	ToggleUserStatus(ctx context.Context, id string, isActive *bool) error
}