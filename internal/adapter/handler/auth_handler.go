package handler

import (
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
// @Router /auth/register [post]
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
// @Router /auth/login [post]
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
