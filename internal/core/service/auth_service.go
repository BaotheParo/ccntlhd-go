package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
	"github.com/yourname/ticketing-system/pkg/auth" // Import bộ công cụ băm và JWT
)

type authService struct {
	userRepo  port.UserRepositoryPort
	secretKey string // Khóa bí mật để ký JWT
}

// NewAuthService khởi tạo Service với Repository và Secret Key
func NewAuthService(repo port.UserRepositoryPort, secretKey string) port.AuthServicePort {
	return &authService{
		userRepo:  repo,
		secretKey: secretKey,
	}
}

// Register xử lý logic Đăng ký
func (s *authService) Register(ctx context.Context, req entity.RegisterRequest) (*entity.User, error) {
	// 1. Kiểm tra email đã tồn tại chưa
	existingUser, _ := s.userRepo.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("email đã được sử dụng")
	}

	// 2. Băm mật khẩu bằng công cụ Bcrypt đã viết ở pkg/auth
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 3. Khởi tạo thực thể User mới (Dùng Constructor như bạn đã hỏi)
	newUser := &entity.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         entity.RoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 4. Gọi Repository để lưu vào Database
	err = s.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

// Login xử lý logic Đăng nhập
func (s *authService) Login(ctx context.Context, req entity.LoginRequest) (string, error) {
	// 1. Tìm User theo email
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return "", errors.New("sai email hoặc mật khẩu")
	}

	// 2. Đối chiếu mật khẩu bằng Bcrypt
	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		return "", errors.New("sai email hoặc mật khẩu")
	}

	// 3. Nếu đúng mật khẩu, tiến hành tạo thẻ JWT (Token)
	token, err := auth.GenerateToken(user.ID.String(), user.Role, s.secretKey)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateToken dùng để kiểm tra thẻ JWT có hợp lệ không
func (s *authService) ValidateToken(ctx context.Context, token string) (*entity.User, error) {
	claims, err := auth.ValidateToken(token, s.secretKey)
	if err != nil {
		return nil, err
	}

	// Lấy lại thông tin User từ DB dựa trên ID trong token (để đảm bảo user vẫn tồn tại)
	return s.userRepo.GetUserByID(ctx, claims.UserID)
}
