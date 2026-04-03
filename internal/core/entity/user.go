package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type User struct {
	ID           uuid.UUID              `json:"id"`
	Username     string                 `json:"username"`
	Email        string                 `json:"email"`
	PasswordHash string                 `json:"-"`
	Role         string                 `json:"role"`
	ProfileData  map[string]interface{} `json:"profile_data"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}
