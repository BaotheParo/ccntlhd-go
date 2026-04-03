package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourname/ticketing-system/internal/core/port"
	"github.com/yourname/ticketing-system/pkg/auth"
)

// AuthMiddleware nhận vào secret key để giải mã JWT và kết nối Repository để check trạng thái
func AuthMiddleware(secret string, userRepo port.UserRepositoryPort) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Lấy giá trị từ Header Authorization
		// Thông thường có dạng: Bearer <chuỗi_token>
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Yêu cầu cần có token xác thực",
			})
		}

		// Tách chữ "Bearer " ra để lấy nguyên chuỗi Token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Định dạng Token không hợp lệ (phải là Bearer <token>)",
			})
		}

		tokenString := tokenParts[1]

		// 2. Kiểm tra Token bằng bộ công cụ pkg/auth chúng ta đã viết
		claims, err := auth.ValidateToken(tokenString, secret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token hết hạn hoặc không hợp lệ",
			})
		}

		// 3. (BẢO MẬT) Check DB trực tiếp ở Middleware để đảm bảo Real-time Ban (Khóa là văng ngay lập tức)
		// Lý do: Nếu chỉ check JWT, user bị khóa vẫn xài được đến khi token hết hạn.
		user, err := userRepo.GetUserByID(c.Context(), claims.UserID)
		if err != nil || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Tài khoản không tồn tại",
			})
		}
		if !user.IsActive {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Tài khoản của bạn đã bị khóa. Vui lòng liên hệ Admin",
			})
		}

		// 4. "Đánh dấu" thông tin User vào Request
		// c.Locals giúp truyền dữ liệu từ Middleware vào Handler chính
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)

		// Cho phép đi tiếp vào Handler tiếp theo
		return c.Next()
	}
}

// AdminMiddleware kiểm tra user có role admin không (phải chạy sau AuthMiddleware)
func AdminMiddleware(c *fiber.Ctx) error {
	role := c.Locals("role")
	if role == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Yêu cầu xác thực trước",
		})
	}

	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Chỉ admin mới có quyền thực hiện thao tác này",
		})
	}

	return c.Next()
}
