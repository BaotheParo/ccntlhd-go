package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourname/ticketing-system/pkg/auth"
)

// AuthMiddleware nhận vào secret key để giải mã JWT
func AuthMiddleware(secret string) fiber.Handler {
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

		// 3. "Đánh dấu" thông tin User vào Request
		// c.Locals giúp truyền dữ liệu từ Middleware vào Handler chính
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)

		// Cho phép đi tiếp vào Handler tiếp theo
		return c.Next()
	}
}
