package handlers

import (
	"log/slog"

	"github.com/XianLinNet/XLNetAccount/internal/model"
	"github.com/gofiber/fiber/v2"
)

func writeError(c *fiber.Ctx, status int, message string) error {
	if message == "" {
		message = "error"
	}
	logRequestError(c, status, message)
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

func logRequestError(c *fiber.Ctx, status int, message string) {
	level := slog.LevelWarn
	if status >= 500 {
		level = slog.LevelError
	}
	slog.LogAttrs(c.Context(), level, "request error",
		slog.String("method", c.Method()),
		slog.String("path", c.Path()),
		slog.Int("status", status),
		slog.String("ip", c.IP()),
		slog.String("detail", message),
	)
}
