package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
)

type AuthHandler struct {
	svc port.AuthServicePort
}

func NewAuthHandler(svc port.AuthServicePort) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register godoc
// @Summary Register new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body entity.RegisterRequest true "Register Payload"
// @Success 201 {object} entity.User
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req entity.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dữ liệu không hợp lệ"})
	}

	user, err := h.svc.Register(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// Login godoc
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body entity.LoginRequest true "Login Payload"
// @Success 200 {object} map[string]string "Returns JWT Token"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req entity.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dữ liệu không hợp lệ"})
	}

	token, err := h.svc.Login(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"token": token})
}

// ListUsers godoc
// @Summary List all users (Admin)
// @Tags Admin User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/users [get]
func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	users, total, err := h.svc.ListUsers(c.Context(), page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":  users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// UpdateUserStatus godoc
// @Summary Ban/Unban user (Admin)
// @Tags Admin User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param payload body entity.UpdateUserStatusRequest true "User Status Payload"
// @Success 200 {object} map[string]string
// @Router /api/v1/admin/users/{id}/status [patch]
func (h *AuthHandler) UpdateUserStatus(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID người dùng không được để trống"})
	}

	var req entity.UpdateUserStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Trạng thái is_active không hợp lệ"})
	}

	err := h.svc.ToggleUserStatus(c.Context(), userID, req.IsActive)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cập nhật trạng thái tài khoản thành công",
	})
}
