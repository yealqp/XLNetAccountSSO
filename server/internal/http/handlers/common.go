package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/xianlin-network/sso-platform/server/internal/model"
)

func writeError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"message": message})
}

func publicUser(user *model.User) fiber.Map {
	return fiber.Map{
		"id":           user.ID,
		"username":     user.Username,
		"display_name": user.DisplayName,
		"email":        user.Email,
		"role":         user.Role,
		"status":       user.Status,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	}
}
