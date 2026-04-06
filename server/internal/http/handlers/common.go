package handlers

import (
	"github.com/XianLinNet/XLNetAccount/internal/model"
	"github.com/gofiber/fiber/v2"
)

func writeError(c *fiber.Ctx, status int, message string) error {
	if message == "" {
		message = "error"
	}
	return c.Status(status).JSON(fiber.Map{
		"code":    status,
		"data":    fiber.Map{},
		"message": message,
	})
}

func writeSuccess(c *fiber.Ctx, status int, data any, message string) error {
	if message == "" {
		message = "success"
	}
	if data == nil {
		data = fiber.Map{}
	}
	return c.Status(status).JSON(fiber.Map{
		"code":    status,
		"data":    data,
		"message": message,
	})
}

func publicUser(user *model.User) fiber.Map {
	return fiber.Map{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"status":     user.Status,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}
