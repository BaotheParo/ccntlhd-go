package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/service"
)

func TestUserBanUnbanFlow(t *testing.T) {
	db := setupDB()

	// Đảm bảo cấu trúc bảng luôn mới nhất (Đặc biệt là cột is_active mới thêm)
	if err := db.AutoMigrate(&entity.User{}); err != nil {
		t.Fatalf("Failed to migrate users table: %v", err)
	}

	// Dọn dẹp Database an toàn
	db.Exec("DELETE FROM users")

	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, "test_secret")

	// 1. Khởi tạo một User bình thường
	normalUserID := uuid.New()
	email := fmt.Sprintf("banneduser-%d@example.com", time.Now().UnixNano())
	username := fmt.Sprintf("banneduser-%d", time.Now().UnixNano())

	normalUser := entity.User{
		ID:           normalUserID,
		Username:     username,
		Email:        email,
		PasswordHash: "hash",
		Role:         entity.RoleUser,
		IsActive:     true, // Mặc định true
	}

	if err := db.Create(&normalUser).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	ctx := context.Background()

	// 2. Chức năng Admin: Bắn Status = False (Khóa Account)
	isBan := false
	err := authSvc.ToggleUserStatus(ctx, normalUserID.String(), &isBan)
	if err != nil {
		t.Fatalf("Failed to BAN user: %v", err)
	}

	fmt.Println("✅ [Admin] Đã ra lệnh khóa Account thành công!")

	// 3. Giả lập luồng Middleware: User tiếp tục cầm Token cũ đi gọi API
	// Middleware sẽ móc ID ra và gọi GetUserByID để trực tiếp kiểm tra DB.
	userInDB, err := userRepo.GetUserByID(ctx, normalUserID.String())
	if err != nil || userInDB == nil {
		t.Fatalf("Failed to fetch user from DB in Middleware simulation")
	}

	if userInDB.IsActive {
		t.Fatalf("Expected User to be BANNED (IsActive=false), but got TRUE")
	}

	fmt.Println("✅ [Middleware] Real-time Ban Check Pass: Phát hiện User bị khóa! (IsActive == false). Sút văng ra 403 Forbidden!")

	// 4. Chức năng Admin: Mở khóa lại (Unban)
	isUnban := true
	err = authSvc.ToggleUserStatus(ctx, normalUserID.String(), &isUnban)
	if err != nil {
		t.Fatalf("Failed to UNBAN user: %v", err)
	}

	fmt.Println("✅ [Admin] Đã ra lệnh Mở khóa Account thành công!")

	// 5. Kiểm tra DB xem kích hoạt lại chưa
	userInDBAfterUnban, _ := userRepo.GetUserByID(ctx, normalUserID.String())
	if !userInDBAfterUnban.IsActive {
		t.Fatalf("Expected User to be UNBANNED (IsActive=true), but got FALSE")
	}

	fmt.Println("✅ [Middleware] Real-time Ban Check Pass: Phát hiện User đã được thả! (IsActive == true). Cho phép truy cập!")
}
